// http://alexmarandon.com/articles/web_widget_jquery/
// run in a function() so that we don't impact other javascript running on the page
(function () {
  var JiveSearch = {
    widget: function (options) {
      this.options = options;
      this.start();
    },
  }

  JiveSearch.widget.prototype.start = function () {
    var options = this.options;
    if (options.host === "") {
      options.host = "https://jivesearch.com";
    }

    if (typeof jQuery == 'undefined' || window.jQuery.fn.jquery !== '1.12.2') {
      function getScript(url, success) {
        var script = document.createElement('script');
        script.src = url;
        var head = document.getElementsByTagName('head')[0],
          done = false;
        script.onload = script.onreadystatechange = function () {
          if (!done && (!this.readyState || this.readyState == 'loaded' || this.readyState == 'complete')) {
            done = true;
            success();
            script.onload = script.onreadystatechange = null;
            head.removeChild(script);
          };
        };
        head.appendChild(script);
      };
      getScript(options.host + '/static/jquery-1.12.2.min.js', function () {
        if (typeof jQuery == 'undefined') {
          // Super failsafe - still somehow failed...
          console.log("jquery did not load...");
        } else {
          $(document).ready(function () {
            main(options);            
          });
        }
      });
    } else { // jQuery was already loaded
      $.noConflict();
      $(document).ready(function () {
        main(options);
      });
    };
  };

  function main(options) {
    if (options.query === "") {
      console.log("please provide a 'query' parameter");
      return
    }
    jQuery.ajax({
      url: options.host + "/answer?q=" + options.query,
      dataType: "JSONP", // JSON results in inconsistent loading of javascript. Only JSONP seems to work??
      jsonpCallback: "jivesearchcallback"
    }).done(function (data) {
      /******* Load CSS *******/
      /* We probably don't need all these... */
      var cssFiles = [options.host + "/static/pure-min.css", options.host + "/static/grids-responsive-old-ie-min.css",
        options.host + "/static/grids-responsive-min.css", options.host + "/static/fonts/css/fonts.css",
        options.host + "/static/fontello/css/fontello.css", options.host + "/static/base.css", options.host + "/static/search.css"
      ];
      var allCSS = cssFiles.concat(data.css);
      jQuery.each(allCSS, function (index, val) {
        jQuery("<link/>", {
          rel: "stylesheet",
          type: "text/css",
          href: val
        }).appendTo("head");
      });

      /******* Load HTML *******/
      jQuery("#jive_search").html(data.html);

      /******* Load JavaScript *******/
      // If > 1 JavaScript file, make sure d3.js and other libraries are fully loaded before we get the last script
      if (data.javascript.length > 1) {
        getMultiScripts = function (arr) {
          var _arr = jQuery.map(arr, function (scr) {
            return jQuery.getScript(scr);
          });
          _arr.push(jQuery.Deferred(function (deferred) {
            jQuery(deferred.resolve);
          }));
          return jQuery.when.apply(jQuery, _arr);
        }
        getMultiScripts(data.javascript.slice(0, -1)).done(function () {
          jQuery("<script/>", {
            src: data.javascript[data.javascript.length - 1]
          }).appendTo("body");
        });
      } else if (data.javascript.length === 1) {
        jQuery("<script/>", {
          src: data.javascript[0]
        }).appendTo("body");
      }

      /******* Styling, etc. *******/
      var src = '<em>Source:</em>';
      src += ' <a href="https://jivesearch.com"><em>Jive Search</em>'
      src += '<span id="jivesearch_src_tagline">&nbsp;- A private search engine</span></a>';
      jQuery("#jive_search #source").html(src);
      jQuery("#jive_search #source a").css("text-decoration", "none");
      jQuery("#jive_search #source a #jivesearch_src_tagline").css("color", "#333")
        .css("display", "inline-block")
        .css("text-decoration", "none")
        .css("cursor", "default");
      jQuery("#jive_search #source a:hover #jivesearch_src_tagline").css("text-decoration", "none");
      jQuery("#jive_search #source a:active #jivesearch_src_tagline").css("text-decoration", "none");
      jQuery("#jive_search #source a:visited #jivesearch_src_tagline").css("text-decoration", "none");

      if (options.border === false) {
        jQuery("#jive_search #answer").css("box-shadow", "none");
      }
      if (options.width != "") {
        jQuery("#jive_search #answer").css("width", options.width);
      }
      if (options.height != "") {
        jQuery("#jive_search #answer").css("height", options.height);
      }
    });
  }

  window.JiveSearch = JiveSearch
})(); // We call our anonymous function immediately

function jivesearchcallback(result) { }
