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
        responsive:{
            0:{
                items:1
            },
            300:{
                items:2 
            },        
            425:{
                items:3
            },
            550:{
                items:4
            },
            675:{
                items:5
            },
            1000:{
                items:6
            },
            1125:{
                items:7
            },
            1250:{
                items:8
            },
            1375:{
                items:9
            },
            1450:{
                items:11
            },
            1575:{
                items:12
            },
            1700:{
                items:13
            },
            1825:{
                items:14
            },
            1950:{
                items:15
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


