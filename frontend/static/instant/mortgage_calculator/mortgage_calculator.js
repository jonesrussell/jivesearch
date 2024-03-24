var yrs;
var rate;
var amt;
var pmt;

window.onload = function(){
  document.getElementById("amt").focus();
  $(".mtg").on("change", function() {
    getValues();
  });
};

function getValues(){
  yrs = document.getElementById("yrs").value;
  rate = document.getElementById("rate").value;
  amt = document.getElementById("amt").value;
  rate /= 1200;
  yrs *= 12;
  pmt = calculatePayment();
  if (!isNaN(parseFloat(pmt)) && isFinite(pmt)){
    $("#pmt").html("$" + pmt.toFixed(2));
  }
};

function calculatePayment(){
	var p = amt*(rate * Math.pow((1 + rate), yrs))/(Math.pow((1 + rate), yrs) - 1);
	return p;
}