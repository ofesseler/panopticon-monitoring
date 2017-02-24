/**
 * Created by oli on 2/11/17.
 */
$(document).ready(function () {

    var healthReq = $.ajax({
        url: "/api/v1/health",
        method: "GET"
    }).done(function(data){
        currentState = $('#current-state');
        cs = $('div#clusterstatus');
        switch (data.ClusterState) {
            case 1:
                currentState.text("healthy");
                cs.css("background-color", "#3bcc1a");
                console.log(cs)
                break;
            case 2:
                currentState.text("warning");
                cs.css("background-color", "#e2b500");
                break;
            case 3:
                currentState.text("critical");
                cs.css("background-color", "#af0726");
                break;
        }
    }).fail(function (jqXHR, textStatus) {
        console.log(textStatus)
    });

    var consulReq = $.ajax({
        url: "/api/v1/consul/health",
        method: "GET"
    }).done(function(data){
        $("#consul-data").text(JSON.stringify(data))
    }).fail(function (jqXHR, textStatus) {
        console.log(textStatus)
    });

    var glusterReq = $.ajax({
        url: "/api/v1/gluster/health",
        method: "GET"
    }).done(function(data){
        $("#gluster-data").text(JSON.stringify(data))
    }).fail(function (jqXHR, textStatus) {
        console.log(textStatus)
    });

    var weaveReq = $.ajax({
        url: "/api/v1/weave/health",
        method: "GET"
    }).done(function(data){
        $("#weave-data").text(JSON.stringify(data))
    }).fail(function (jqXHR, textStatus) {
        console.log(textStatus)
    });

});