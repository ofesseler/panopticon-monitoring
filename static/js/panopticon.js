/**
 * Created by oli on 2/11/17.
 */
$(document).ready(function () {
    var color_healthy   = "#00ae00";
    var color_warning  = "#fdd00c";
    var color_critical = "#dc2300";

    function changeDivColor(textStatusID, wrapperID, status) {
        switch (status) {
            case 1:
                textStatusID.text("healthy");
                wrapperID.css("background-color", color_healthy);
                break;
            case 2:
                textStatusID.text("warning");
                wrapperID.css("background-color", color_warning);
                break;
            case 3:
                textStatusID.text("critical");
                wrapperID.css("background-color", color_critical);
                break;
        }
    }

    var healthReq = $.ajax({
        url: "/api/v1/health",
        method: "GET"
    }).done(function(data){
        currentState = $('#current-state');
        cs = $('div#clusterstatus');
        changeDivColor(currentState, cs, data.ClusterState)
    }).fail(function (jqXHR, textStatus) {
        console.log(textStatus)
    });

    var consulReq = $.ajax({
        url: "/api/v1/consul/health",
        method: "GET"
    }).done(function(data){
        consulData = $("#consul-data")
        consulData.text(JSON.stringify(data))
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

    var hostsReq = $.ajax({
        url: "/api/v1/hosts/health",
        method: "GET"
    }).done(function(data){
        $("#hosts-data").text(JSON.stringify(data))
    }).fail(function (jqXHR, textStatus) {
        console.log(textStatus)
    });

});
