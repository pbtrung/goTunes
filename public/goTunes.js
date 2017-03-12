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
            getItemQueryResults(query, function(results) {
                var models = _.map(results, function(d) {
                    return new Item(d);
                });
                var rlts = new Items(models);
                app.showResults(rlts);
            });
        }
    });
    var router = new Router();
    Backbone.history.start();   

    var AppView = Backbone.View.extend({
        el: $("body"),
        events: {
            "submit #queryForm": "routeResults"
        },

        routeResults: function(ev) {
            ev.preventDefault();
            router.navigate("search/" + encodeURIComponent($("#query").val()), { trigger: true });
        },
        showResults: function(results) {
            var source = $("#template").html();
            var template = Handlebars.compile(source);
            var html;
            $("#results").empty();
            results.each(function(result) {
                html = template(result.toJSON());
                $("#results").append(html);
            });
        }
    });

    var app = new AppView();
});

function getItemQueryResults(query, callback) {
    if(localStorage.getItem("awsToken") === null) {
        getTokenSearch(query, callback);
    } else {
        var awsToken = JSON.parse(localStorage.getItem("awsToken"));
        var now = new Date();
        var tokenDate = new Date(token.Expiration);
        if(now > tokenDate) {
            getTokenSearch(query, callback);
        } else {
            search(awsToken, query, callback);
        }
    }
}

function getTokenSearch(query, callback) {
    $.ajax({
        type: "GET",
        url: "/awsToken",
        dataType: "json",
        success: function(token) {
            localStorage.setItem("awsToken", JSON.stringify(awsToken));
            search(awsToken, query, callback);
        }
    });
}

function search(awsToken, query, callback) {
    AWS.config.credentials = new AWS.Credentials(awsToken.AccessKeyID, awsToken.SecretAccessKey, awsToken.SessionToken);
    AWS.config.update({ region: awsToken.Region });
    var dynamodb = new AWS.DynamoDB();

    var params = {
        ExpressionAttributeValues: {
            ":q": { S: query }
        }, 
        KeyConditionExpression: "title CONTAINS :q", 
        ProjectionExpression: "id, album, albumartist, title, track, tracktotal, format, length, bitrate, mb_trackid, lyrics, fileid, albumartid", 
        Limit: 6,
        ConsistentRead: false,
        TableName: "music"
    };
    dynamodb.query(params, function(err, data) {
        if(err) {
            console.log(err, err.stack);
        } else {
            var items = [];
            for(var i in data.Items) {
                var item = {};
                item.id = i.id.N;
                item.album = i.album.S;
                item.albumartist = i.albumartist.S;
                item.title = i.title.S;
                item.track = i.track.N;
                item.tracktotal = i.tracktotal.N;
                item.format = i.format.S;
                item.length = timeFormat(i.length.N);
                item.bitrate = Math.round(i.bitrate.N / 1000);
                item.mb_trackid = i.mb_trackid.S;
                item.lyrics = i.lyrics.S;
                item.fileid = i.fileid.S;
                item.albumartid = i.albumartid.S;
                items.push(item);
            }
            callback(items);
        }          
    });
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
