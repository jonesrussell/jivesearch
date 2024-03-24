#!/usr/bin/env python
"""
The following combines Yahoo's open_nsfw and TensorFlow's imagenet:
https://github.com/yahoo/open_nsfw/blob/master/classify_nsfw.py
https://github.com/tensorflow/models/blob/master/tutorials/image/imagenet/classify_image.py
"""
from __future__ import absolute_import
from __future__ import division
from __future__ import print_function
from bottle import route, request, response, run
from gevent import monkey; monkey.patch_all()
from json import dumps
from cgi import parse_qs
import numpy as np
from PIL import Image
from PIL.ExifTags import TAGS
import requests
from StringIO import StringIO
import os.path
import re
import sys
import tarfile, zipfile
from six.moves import urllib
import magic
import tensorflow as tf
import caffe

# pylint: disable=line-too-long
# pylint: enable=line-too-long

def nsfw_score(im, caffe_transformer=None, caffe_net=None, output_layers=None):
  d = {}

  if caffe_net is None:
    return d

  if output_layers is None:
    output_layers = caffe_net.outputs
  
  """
  It is important to use this resize logic as it was 
  used to generate the training dataset
  """ 
  if im.mode != "RGB":
      im = im.convert('RGB')
  imr = im.resize((256, 256), resample=Image.BILINEAR)
  fh_im = StringIO()
  imr.save(fh_im, format='JPEG')
  fh_im.seek(0)
  img_data_rs = bytearray(fh_im.read())

  image = caffe.io.load_image(StringIO(img_data_rs))
  H, W, _ = image.shape
  _, _, h, w = caffe_net.blobs['data'].data.shape
  h_off = max((H - h) // 2, 0)
  w_off = max((W - w) // 2, 0)
  crop = image[h_off:h_off + h, w_off:w_off + w, :]
  transformed_image = caffe_transformer.preprocess('data', crop)
  transformed_image.shape = (1,) + transformed_image.shape

  input_name = caffe_net.inputs[0]
  all_outputs = caffe_net.forward_all(blobs=output_layers,
    **{input_name: transformed_image})

  # outputs[1] indicates the NSFW probability
  outputs = all_outputs[output_layers[0]][0].astype(float)
  d["nsfw_score"] = np.float64(outputs[1])
  return d

def metadata(im):
  d = {}
  d["width"], d["height"] = im.size

  if hasattr(im, "_getexif"):
    info = im._getexif()
    if info:
      for tag, value in info.items():
        try:
          decoded = TAGS.get(tag, tag).lower()
        except AttributeError:
          decoded = TAGS.get(tag, tag)
        d[decoded] = value
  
  return d

class NodeLookup(object):
  """Converts integer node ID's to human readable labels."""
  def __init__(self, label_lookup_path=None, uid_lookup_path=None):
    if not label_lookup_path:
      label_lookup_path = 'imagenet_2012_challenge_label_map_proto.pbtxt'
    if not uid_lookup_path:
      uid_lookup_path = 'imagenet_synset_to_human_label_map.txt'
    self.node_lookup = self.load(label_lookup_path, uid_lookup_path)

  def load(self, label_lookup_path, uid_lookup_path):
    """Loads a human readable English name for each softmax node.

    Args:
      label_lookup_path: string UID to integer node ID.
      uid_lookup_path: string UID to human-readable string.

    Returns:
      dict from integer node ID to human-readable string.
    """
    if not tf.gfile.Exists(uid_lookup_path):
      tf.logging.fatal('File does not exist %s', uid_lookup_path)
    if not tf.gfile.Exists(label_lookup_path):
      tf.logging.fatal('File does not exist %s', label_lookup_path)

    # Loads mapping from string UID to human-readable string
    proto_as_ascii_lines = tf.gfile.GFile(uid_lookup_path).readlines()
    uid_to_human = {}
    p = re.compile(r'[n\d]*[ \S,]*')
    for line in proto_as_ascii_lines:
      parsed_items = p.findall(line)
      uid = parsed_items[0]
      human_string = parsed_items[2]
      uid_to_human[uid] = human_string

    # Loads mapping from string UID to integer node ID.
    node_id_to_uid = {}
    proto_as_ascii = tf.gfile.GFile(label_lookup_path).readlines()
    for line in proto_as_ascii:
      if line.startswith('  target_class:'):
        target_class = int(line.split(': ')[1])
      if line.startswith('  target_class_string:'):
        target_class_string = line.split(': ')[1]
        node_id_to_uid[target_class] = target_class_string[1:-2]

    # Loads the final mapping of integer node ID to human-readable string
    node_id_to_name = {}
    for key, val in node_id_to_uid.items():
      if val not in uid_to_human:
        tf.logging.fatal('Failed to locate: %s', val)
      name = uid_to_human[val]
      node_id_to_name[key] = name

    return node_id_to_name

  def id_to_string(self, node_id):
    if node_id not in self.node_lookup:
      return ''
    return self.node_lookup[node_id]

def classify_image(image_data):
  d = {}

  with tf.Session() as sess:
    # Some useful tensors:
    # 'softmax:0': A tensor containing the normalized prediction across
    #   1000 labels.
    # 'pool_3:0': A tensor containing the next-to-last layer containing 2048
    #   float description of the image.
    # 'DecodeJpeg/contents:0': A tensor containing a string providing JPEG
    #   encoding of the image.
    # Runs the softmax tensor by feeding the image_data as input to the graph.
    softmax_tensor = sess.graph.get_tensor_by_name('softmax:0')
    predictions = sess.run(softmax_tensor, {'DecodeJpeg/contents:0': image_data})
    predictions = np.squeeze(predictions)

    # Creates node ID --> English string lookup.
    node_lookup = NodeLookup()

    top_k = predictions.argsort()[-5:][::-1]
    for node_id in top_k:
      human_string = node_lookup.id_to_string(node_id)
      score = predictions[node_id]
      d[human_string] = np.float64(score)
  
  return d

@route('/')
def index():
  print(request.query.image)
  response.content_type = 'application/json'

  headers = {
    'User-Agent':
      os.environ.get('JIVESEARCH_CRAWLER_USERAGENT_FULL',
      'https://github.com/jivesearch/jivesearch'),
  }

  try:
    image_data = requests.get(request.query.image, headers=headers).content
    im = Image.open(StringIO(str(image_data)))
  except Exception, e:
    print(e)
    return {}

  # classify_image_graph_def.pb:
  #   Binary representation of the GraphDef protocol buffer.
  # imagenet_synset_to_human_label_map.txt:
  #   Map from synset ID to a human readable string.
  # imagenet_2012_challenge_label_map_proto.pbtxt:
  #   Text representation of a protocol buffer mapping a label to synset ID.
  classification = classify_image(image_data)

  nsfw = nsfw_score(im, caffe_transformer=caffe_transformer, 
      caffe_net=nsfw_net, output_layers=['prob'],
  )

  md = metadata(im)

  d = nsfw.copy()
  d.update(md)
  d["classification"] = classification
  d["mime"] = magic.from_buffer(image_data, mime=True)

  try:
    return dumps(d)
  except Exception, e:
    print(e)
    return dumps({})

nsfw_net = caffe.Net("nsfw_model/deploy.prototxt", "nsfw_model/resnet_50_1by2_nsfw.caffemodel", caffe.TEST)
caffe_transformer = caffe.io.Transformer({'data': nsfw_net.blobs['data'].data.shape})
caffe_transformer.set_transpose('data', (2, 0, 1))  # move image channels to outermost
caffe_transformer.set_mean('data', 
  np.array([104, 117, 123]))  # subtract the dataset-mean value in each channel
caffe_transformer.set_raw_scale('data', 255)  # rescale from [0, 1] to [0, 255]
caffe_transformer.set_channel_swap('data', (2, 1, 0))  # swap channels from RGB to BGR

# TensorFlow
with tf.gfile.FastGFile('classify_image_graph_def.pb', 'rb') as f:
  graph_def = tf.GraphDef()
  graph_def.ParseFromString(f.read())
  _ = tf.import_graph_def(graph_def, name='')

if __name__ == '__main__':
  p = int(os.environ.get('NSFW_PORT','8080'))
  print('Serving on ', p)
  run(host=os.environ.get('NSFW_ADDR','127.0.0.1'), port=p, server='gunicorn')