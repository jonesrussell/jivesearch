$(document).ready(function() {
    $("#live").show();

    $("#live").on("click", function(){
        alert("clicked");
        $(".live").show();
    });

    $('.owl-carousel').owlCarousel({
        loop: false,
        nav: true,
        navText: [
            "<i class='icon-left-open-mini' style='display:none;'></i>",
            "<i class='icon-right-open-mini'></i>"
        ],
        autoplay: true,
        autoplayHoverPause: true,
        responsive:{ // e.g. 1950 / 175 width of each = 11.14 -> round down to 11.
            0:{
                items:1
            },
            300:{
                items:1
            },        
            425:{
                items:2
            },
            550:{
                items:3
            },
            675:{
                items:3
            },
            1000:{
                items:5
            },
            1125:{
                items:6
            },
            1250:{
                items:7
            },
            1375:{
                items:7
            },
            1450:{
                items:8
            },
            1575:{
                items:8
            },
            1700:{
                items:8
            },
            1825:{
                items:9
            },
            1950:{
                items:10
            }
        }
    });
    
    var owl = $('.owl-carousel');
    owl.owlCarousel();
    // Listen to owl events:
    owl.on('changed.owl.carousel', function(event) {
        $(".icon-left-open-mini").show();
    });

});