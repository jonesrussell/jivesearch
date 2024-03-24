window.onload = function() {
    // hide the non-js answer and show calculator
    $("#calculator").show();

    var output;
    var result = document.getElementById("result");
    var numbers = document.querySelectorAll(".number");

    // numbers
    function enterNumber(num){
        output = result.innerHTML += num;
    }
    for(var i = 0; i < numbers.length; i++ ) {
        numbers[i].addEventListener("click", function() {
            enterNumber(this.value);
        }, false);        
    } 

    // "0"
    function enterZero(){
        if(result.innerHTML === "" || result.innerHTML === "0") { // don't want leading zeroes..."0345" => "345"
            output = result.innerHTML = "0";  
        } else if(result.innerHTML === output) {
            output = result.innerHTML += "0";
        }
    }
    document.querySelector(".zero").addEventListener("click", function() {
        enterZero();
    }, false);
        
    // "."
    function enterPeriod(){
        if(result.innerHTML === "0."){
            // do nothing
        } else if(result.innerHTML === "") {
            output = result.innerHTML = result.innerHTML.concat("0.");
        } else if(result.innerHTML === output) {
            result.innerHTML = result.innerHTML.concat(".");
        }
    }
    document.querySelector(".period").addEventListener("click", function() {
        enterPeriod();
    }, false);
    
    // "="
    function enterEqual(){
        if(result.innerHTML === output) {
            result.innerHTML = eval(output);
        } else {
            result.innerHTML = "";
        }
    }
    document.querySelector("#eqn-bg").addEventListener("click",function() {
        enterEqual();
    }, false);
        
    // clear
    document.querySelector("#clear").addEventListener("click",function() {
        result.innerHTML = "";
    }, false);

    // press "+", "-", "*", "/"
    var operators = document.querySelectorAll(".operator");

    for(var i = 0; i < operators.length; i++ ) {
        operators[i].addEventListener("click",function() {
            enterOperator(this.value);
        },false);
    }

    function enterOperator(val){
        if(result.innerHTML === "") {
            result.innerHTML = result.innerHTML.concat("");
        } else if(output) {
            result.innerHTML = output.concat(val);
        }
    }

    function backspaceOperator(val){
        result.innerHTML = output = result.innerHTML.slice(0, -1);
        if (output.slice(-1) in ops){
            output = output.slice(0, -1);
        }
    }

    // listen to keyboard   
    function act(event) {
        var key = 0;
        if (window.event) {
            key = window.event.keyCode;
        } else if (event) {
            key = event.keyCode;
        }
        
        if (key===48 || key===96){ // 0
            enterZero();
        } else if (key===110 || key===190){ // .
            enterPeriod();
        } else if (key in keyCodeNumbers){ // 1-9
            enterNumber(keyCodeNumbers[key]);
        } else if (key===187){
            if (event.shiftKey) { // +
                enterOperator("+");
            } else { // =
                enterEqual();
            }
        } else if (key===18){
            if (event.shiftKey) { // *
                enterOperator("*");
            }
        } else if (key===8){ // backspace
            backspaceOperator();
        } else if (key===13){ // enter
            enterEqual();
        } else if (key in keyCodeOperators){ // + - * =/
            enterOperator(keyCodeOperators[key]);
        }

        return true;
    }

    var active=document.getElementById("result");
    active.onkeydown=act;
}

var keyCodeNumbers = {
    49:"1",50:"2",51:"3",52:"4",53:"5",54:"6",55:"7",56:"8",57:"9",
    97:"1",98:"2",99:"3",100:"4",101:"5",102:"6",103:"7",104:"8",105:"9", /* numpad */
};

var keyCodeOperators = {
    189:"-",191:"/",
    109:"-",111:"/",106:"*",107:"+" /* numpad */
};

var ops = {"+":1, "-":1, "*":1, "/":1};

// % sign
/*
var NUMBER_FIVE = 53;
element.onkeydown = function (event) {
    if (event.keyCode == NUMBER_FIVE) {
        if (event.shiftKey) {
            // '%' handler
        } else {
            // '5' handler
        }
    }
};
*/

        