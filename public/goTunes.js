// Format times as minutes and seconds.
var timeFormat = function(secs) {
    if (secs == undefined || isNaN(secs)) {
        return "0:00";
    }
    secs = Math.round(secs);
    var mins = "" + Math.round(secs / 60);
    secs = "" + secs % 60;
    if (secs.length < 2) {
        secs = "0" + secs;
    }
    return mins + ":" + secs;
};

$(function() {

    // Model
    var Item = Backbone.Model.extend({
    });
    var Items = Backbone.Collection.extend({
        model: Item
    });

    var Router = Backbone.Router.extend({
        routes: {
            "search/:query": "itemQuery"
        },
        itemQuery: function(query) {
            if(localStorage.getItem("awsToken") === null) {
                $.ajax({
                    type: "GET",
                    url: "/awsToken",
                    dataType: "json",
                    success: function(token) {
                        localStorage.setItem("awsToken", JSON.stringify(awsToken));
                        search(awsToken, query);
                    }
                });
            } else {
                var awsToken = JSON.parse(localStorage.getItem("awsToken"));
                var now = new Date();
                var tokenDate = new Date(token.Expiration);
                if(now >= tokenDate) {
                    $.ajax({
                        type: "GET",
                        url: "/awsToken",
                        dataType: "json",
                        success: function(awsToken) {
                            localStorage.setItem("awsToken", JSON.stringify(awsToken));
                        }
                    });
                } else {

                }
            }
        }
    });
    var router = new Router();
    Backbone.history.start();   

    var AppView = Backbone.View.extend({
        el: $("body"),
        events: {
            "submit #queryForm": "routeResults",
        },

        routeResults: function(ev) {
            ev.preventDefault();
            router.navigate("search/" + encodeURIComponent($("#query").val()), { trigger: true });
        }
    });

    var app = new AppView();
});

function search(awsToken, query) {
    AWS.config.credentials = new AWS.Credentials(awsToken.AccessKeyID, awsToken.SecretAccessKey, awsToken.SessionToken);
    AWS.config.update({ region: "us-west-2" });
    var dynamodb = new AWS.DynamoDB();
}

function getItemURL(fileID, callback) {
    if(localStorage.getItem("gdrToken") === null) {
        getItemURLFromGDR(fileID, callback);
    } else {
        var gdrToken = JSON.parse(localStorage.getItem("gdrToken"));
        var now = new Date();
        var tokenDate = new Date(gdrToken.Expiration);
        if(now > tokenDate) {
            getItemURLFromGDR(fileID, callback);
        } else {
            getItemURLFromGDRWithToken(fileID, gdrToken, callback); 
        }
    }
}

function getItemURLFromGDR(fileID, callback) {
    $.ajax({
        type: "GET",
        url: "/gdrToken",
        dataType: "json",
        success: function(gdrToken) {
            localStorage.setItem("gdrToken", JSON.stringify(gdrToken));
            getItemURLFromGDRWithToken(fileID, gdrToken, callback);
        }
    });
}

function getItemURLFromGDRWithToken(fileID, gdrToken, callback) {
    $.ajax({
        type: "GET",
        url: "https://www.googleapis.com/drive/v2/files/" + fileID,
        headers: {
            Authorization: "Bearer " + gdrToken.AccessToken
        },
        dataType: "json",
        success: function(item) {
            var itemURL = item.downloadUrl + "&access_token=" + encodeURIComponent(gdrToken.AccessToken);
            callback(itemURL);
        }
    });
}
