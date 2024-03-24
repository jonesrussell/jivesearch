$(document).ready(function(){
    $("#answer").css("height", "400px");
    draw("5yr");

    function draw(id){
        var len;

        $("#stock_chart").html(""); // we can work on a transition another day

        var date = new Date();
        var xFormat = d3.timeFormat("%b %e");

        if (id==="day"){
            
        }else if (id==="week"){
            date.setDate(date.getDate() - 7);
        }else if (id==="month"){
            date.setMonth(date.getMonth() - 1);
        }else if (id==="ytd"){
            date.setFullYear(date.getFullYear() - 1);
            date.setMonth(11);
            date.setDate(31);
            xFormat = d3.timeFormat("%b-%y");          
        }else if (id==="1yr"){
            date.setFullYear(date.getFullYear() - 1);
            xFormat = d3.timeFormat("%b-%y");
        }else if (id==="5yr"){
            date.setFullYear(date.getFullYear() - 5);
            xFormat = d3.timeFormat("%Y");
        }

        var parseTime = d3.timeParse("%Y-%m-%dT00:00:00Z"); // 2018-01-12T00:00:00Z
        var tmp = JSON.parse(JSON.stringify(data));
        tmp = tmp.filter(function(e){
            return parseTime(e.date) >= date; // 1 Week chart doesn't show 5 trading days...
        });
        
        var xTicks = 5;
        if (tmp.length < xTicks){
            xTicks = tmp.length;
        }

        var margin = {top: 20, right: 70, bottom: 30, left: 50},
            width = 530 - margin.left - margin.right, 
            height = 180 - margin.top - margin.bottom; 

        var plotHeight = height - margin.top - margin.bottom;
        
        var toolTipTimeFormat = d3.timeFormat("%a %B %e, %Y");
        var formatNumber = d3.format(",");
        var bisectDate = d3.bisector(function(d){ 
            return d.date; 
        }).left;

        var x = d3.scaleTime().range([0, width]);
        var y = d3.scaleLinear().range([height, 0]);
        var y1 = d3.scaleLinear().range([height, 0]);

        var line = d3.line().x(function(d){ return x(d.date); }).y(function(d){ return y(d.close); });
        var area = d3.area().x(function(d){ return x(d.date); }).y1(function(d){ return y(d.close); });
        var div = d3.select("#stock_chart").append("div").attr("class", "tooltip").style("opacity", 0);
        var svg = d3.select("#stock_chart").append("svg").attr("width", width + margin.left + margin.right)
                    .attr("height", height + margin.top + margin.bottom)
                    .append("g").attr("transform","translate(" + margin.left + "," + margin.top + ")");

        tmp.forEach(function(d){
            d.date = parseTime(d.date);
            d.close = +d.close;
        });

        x.domain(d3.extent(tmp, function(d){ return d.date; }));
        y.domain(d3.extent(tmp, function(d){ return d.close; }));
        y1.domain([0, d3.max(tmp, function(d){ return d.close; })]);
        area.y0(y1(0));

        svg.append("path").datum(tmp).attr("class", "line").attr("d", area);
        svg.append("g").call(d3.axisBottom(x).ticks(xTicks).tickFormat(xFormat))
            .attr("class", "axis").attr("transform", "translate(0," + height + ")");
        svg.append("g").call(d3.axisLeft(y).ticks(5)).attr("class", "axis");
        svg.append("g").call(d3.axisRight(y1)).attr("class", "rightAxis").attr("transform", "translate( " + width + ", 0 )");

        var focus = svg.append("g").attr("class", "focus").style("display", "none");
        focus.append("line").attr("class", "x-hover-line hover-line").attr("y1", 0).attr("y2", height);
        focus.append("line").attr("class", "y-hover-line hover-line").attr("x1", width).attr("x2", width);
        focus.append("circle").attr("r", 1.5);
        focus.append("text").attr("x", 15).attr("dy", ".31em");

        svg.append("rect")
            .attr("transform", "translate(" + 1 + "," + 1 + ")")
            .attr("class", "overlay").attr("width", width).attr("height", height)
            .on("mouseover", function(d){ 
                focus.style("display", null);
            })
            .on("mouseout", function(){ 
                div.style("opacity", 0);
                focus.style("display", "none"); 
            })
            .on("mousemove", mousemove);

        function mousemove(){
            var x0 = x.invert(d3.mouse(this)[0]),
                i = bisectDate(tmp, x0, 1),
                d0 = tmp[i - 1],
                d1 = tmp[i],
                d = x0 - d0.date > d1.date - x0 ? d1 : d0;
            focus.attr("transform", "translate(" + x(d.date) + "," + y(d.close) + ")");
            focus.select(".x-hover-line").attr("y2", height - y(d.close));
            focus.select(".y-hover-line").attr("x2", width + width);
            focus.select("text").text(function(){ 
                div.html("<em>" + formatNumber(d.close) + "</em> " + toolTipTimeFormat(d.date))
                    .style("left", (d3.event.pageX) + "px").style("top", (d3.event.pageY - 28) + "px");
                div.style("opacity", .9);
                return;
            });
        }
    }

    $("#time_period_buttons > button").on("click", function(){
        var id =  $(this).prop('id');
        draw(id);     
    });
});
