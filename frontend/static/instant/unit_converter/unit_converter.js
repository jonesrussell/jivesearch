window.onload = function() {
    var inputs = $(".unit > input");
    var selects = $(".unit > select");

    $("#unit_converter #selector").change(function() {
        // clear the text entry
        inputs.each(function(){
            $(this).val("");
        });

        // load the new options
        var idx = $("#unit_converter #selector option:selected").index();
        var options = [
            data_storage, length, power, speed, temperature
        ]        

        $(".unit > select").each(function(i) {
            var $el = $(this);
            $el.empty();
            $.each(options[idx], function(key,value) {
                var option = $("<option></option>").attr("value", value).text(key);
                $el.append(option);
            });
            // select the first item in the "left" box and second item in the "second" box
            $el.prop("selectedIndex", i);
        });
    });

    $("#unit_converter #selector").change();

    $(inputs).each(function(i) {
        $(this).on('input', function(){
            if (!isNaN(inputs[i].value)){
                // change the opposite input value
                var j = ((i===0) ? 1 : 0);
                inputs[j].value = convert(i);
            }
        });
    });

    $(selects).each(function(i) {
        // Selecting a different unit should ALWAYS
        // change the second input box. Never the first.
        $(this).change(function() {
            if (!isNaN(inputs[i].value)){
                inputs[1].value = convert(0);
            }
        });
    });

    function convert(i){
        var val = inputs[i].value; // value we are converting FROM
        var units = selects[i].value; // units we are converting FROM
        
        var j = ((i===0) ? 1 : 0);
        var units2 = selects[j].value; // units we are converting TO

        var val2;

        // since f to c is a formula need to treat it differently than others
        if (units === "fahrenheit"){
            val2 = val;
            if (units2 === "celsius"){
                val2 = (val - 32) * 5/9;
            }
        }else if (units === "celsius"){
            val2 = val;
            if (units2 === "fahrenheit"){
                val2 = (val * 9/5) + 32;
            }
        }else{
            val2 = (val * units) / units2;
        }

        return val2;
    }
}


// Select options
/*
    https://en.wikipedia.org/wiki/Bit
    https://en.wikipedia.org/wiki/Byte
    1 byte = 8 bits
    Kilobit = 1000 bits
    Kilobyte = 1000 bytes
    Kibibit = 1024 bits
    Kibibyte = 1024 bytes
    ...
    Units below are in Bits
*/
var data_storage = {
    "Bit":      "1",
    "Byte":     "8",
    "Kilobit":  "1000",
    "Kibibit":  "1024",
    "Kilobyte": "8000",
    "Kibibyte": "8192",
    "Megabit":  "1000000",
    "Mebibit":  "1048576",
    "Megabyte": "8000000",
    "Mebibyte": "8388608",
    "Gigabit":  "1000000000",
    "Gibibit":  "1073741824",
    "Gigabyte": "8000000000",
    "Gibibyte": "8589934592",
    "Terabit":  "1000000000000",
    "Tebibit":  "1099511627776",
    "Terabyte": "8000000000000",
    "Tebibyte": "8796093022208",
    "Petabit":  "1000000000000000",
    "Pebibit":  "1125899906842624",
    "Petabyte": "8000000000000000",
    "Pebibyte": "9007199254740990",
};

// Units below are in Nanometers
var length = {
    "Inch":          "25400000",
    "Feet":          "304800000",
    "Yard":          "914400000",
    "Mile":          "1609347219000",
    "Nanometer":     "1",
    "Micrometer":    "1000",
    "Millimeter":    "1000000",
    "Centimeter":    "10000000",
    "Meter":         "1000000000",
    "Kilometer":     "1000000000000",   
    "Nautical Mile": "1852300000000", // Is this as precise as it gets for this one???
}

// Units below are in Watts
var power = {
    "Watt":       "1", 
    "Kilowatt":   "1000", 
    "Megawatt":   "1000000", 
    "Gigawatt":   "1000000000",
    "Terawatt":   "1000000000000",
    "Petawatt":   "1000000000000000",
    "Exawatt":    "1000000000000000000",
    "Horsepower": "745.69987158227022",
}

// Units below are in Kilometers/Hour
var speed = {
    "Feet/Second":     "1.09728", // 3.6*.3048
    "Miles/Hour":      "1.609344", // 1.609344
    "Meters/Second":   "3.6", // 1*60*60/1000
    "Kilometers/Hour": "1",
    "Knot":            "1.852",
}

// Units below are in celsius
var temperature = {
    "Fahrenheit":          "fahrenheit",
    "Celsius":          "celsius",     
}