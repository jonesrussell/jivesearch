$(document).ready(function(){
  $("#nutrition_header_nutrient_label").html(data.nutrient_label.toTitleCase());
  $("#nutrition_nutrient_units").html(data.nutrient_units);

  var items = $("#nutrition_items");
  
  $(data.items).each(function(i) {
    items.append($("<option>").data("value", this.name).data("nutrient", this.trigger).text(this.name));
  });

  $("#nutrition_header_item").html($(items).find(":selected").data("value"));

  $(items).on("change", function(index) {
    $("#nutrition_header_item").html($(this).find(":selected").data("value"));
    var selected_index = $("#nutrition_serving").prop("selectedIndex");
    loadQuantities(data.items[$(this).prop("selectedIndex")].quantities, selected_index);
  });

  loadQuantities(data.items[$(items).prop("selectedIndex")].quantities, 1);

  $("#nutrition_serving").on("change", function(index) {
    findSelectedValue($(this));
  });
});

String.prototype.toTitleCase = function () {
  return this.replace(/\w\S*/g, function(txt){return txt.charAt(0).toUpperCase() + txt.substr(1).toLowerCase();});
};

function loadQuantities(quantities, selected_index){
  // load the servings for that item
  var servings = $("#nutrition_serving");
  servings.empty();

  $(quantities).each(function(j) {
    var opt = $("<option>");
    if (j==selected_index){
      opt = $("<option selected=true>");
    }
    opt.data("value", this.value);
    opt.data("nutrient", this.nutrient);

    var label = this.quantity_label;
    if (this.quantity){
      label = this.quantity + " " + label;
    }

    if ((this.equivalent) && (this.units)){
      label += " (" + this.equivalent + " " + this.units + ")";
    }     

    servings.append(opt.text(label));
  });

  findSelectedValue(servings);
}

function findSelectedValue(servings){
  var val = $(servings).find(":selected").data("value");
  $("#nutrition_nutrient_value").html(val);
}