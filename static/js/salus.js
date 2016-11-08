url = "api/consul/up";

$(document).ready(function () {
    // shows json request data
    $.getJSON(url, function (data) {
        $("#json_raw").text(JSON.stringify(data, null, 2));
    });


    var radius = 40;
    var distance = radius * 5;
    var strength = 0.5;
    var svg = d3.select("svg"),
        width = +svg.attr("width"),
       height = +svg.attr("height");

    //var color_nodes = d3.scaleOrdinal(d3.schemeCategory10);
    color_nodes = function(x) {
        if (x == 0) {
            return "#ba0101"
        } else if (x == 1) {
            return "#09a514"
        } else if (x == 2) {
            return "#09a514"
        }
        return "#fff"
    };
    color_edges = function (x) {
        if (x == 0) {
            return "#ba0101"
        } else if (x == 1) {
            return "#d86f00"
        } else if (x == 2) {
            return "#09a514"
        }
        return "#fff"
    };

    var simulation = d3.forceSimulation()
        .force("link", d3.forceLink().id(function (d) {
            return d.instance;
        })
            .distance(distance)
            .strength(strength))
        .force("charge", d3.forceManyBody())
        .force("center", d3.forceCenter(width / 2, height / 2));

    d3.json(url, function (error, graph) {
        if (error) throw error;

        var link = svg.append("g")
            .attr("class", "links")
            .selectAll("line")
            .data(graph.links)
            .enter().append("line")
            .attr("stroke-width", function (d) {
                return Math.sqrt(d.value) * 4;
            })
            .attr("stroke", function (d) {
                return color_edges(d.value);
            });

        var node = svg.append("g")
            .attr("class", "nodes")
            .selectAll("circle")
            .data(graph.nodes)
            .enter().append("circle")
            .attr("r", radius)
            .attr("fill", function (d) {
                return color_nodes(d.group);
            })
            .call(d3.drag()
                .on("start", dragstarted)
                .on("drag", dragged)
                .on("end", dragended));

        var text = svg.append("g")
            .attr("class", "texts")
            .selectAll("text")
            .data(graph.nodes)
            .enter().append("text")
            .attr("class", "instance name")
            .text(function (d) {
                return d.instance
            })

        simulation
            .nodes(graph.nodes)
            .on("tick", ticked);

        simulation.force("link")
            .links(graph.links);

        function ticked() {
            link
                .attr("x1", function (d) {
                    return d.source.x;
                })
                .attr("y1", function (d) {
                    return d.source.y;
                })
                .attr("x2", function (d) {
                    return d.target.x;
                })
                .attr("y2", function (d) {
                    return d.target.y;
                });

            node
                .attr("cx", function (d) {
                    return d.x;
                })
                .attr("cy", function (d) {
                    return d.y;
                });

            text
                .attr("x", function (d) {
                    return d.x;
                })
                .attr("y", function (d) {
                    return d.y;
                });
        }
    });

    function dragstarted(d) {
        if (!d3.event.active) simulation.alphaTarget(0.3).restart();
        d.fx = d.x;
        d.fy = d.y;
    }

    function dragged(d) {
        d.fx = d3.event.x;
        d.fy = d3.event.y;
    }

    function dragended(d) {
        if (!d3.event.active) simulation.alphaTarget(0);
        d.fx = null;
        d.fy = null;
    }

});