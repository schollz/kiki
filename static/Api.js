
var KiKiApi = function() {}

KiKiApi.prototype.fetch = function(ajaxObj){
    $.ajax(ajaxObj);
}

KiKiApi.prototype.onError = function(toast_msg, callback){
    return function(xhr, status, err) {
        if (xhr.responseJSON.error) {
            err = xhr.responseJSON.error;
        }
        toast_msg && toastr.error(err, toast_msg || 'Error');
        if (!err.stack) {
            err = new Error(err)
        }
        callback && callback(err);
    }
}

KiKiApi.prototype.onSuccess = function(toast_msg, callback){
    return function(data, status, xhr) {
        if ("ok" != data.status) {
            toast_msg && toastr.error(data.error, toast_msg || 'Error');
            callback && callback(new Error(data.error));
        }
        toast_msg && toastr.success(data.message, toast_msg);
        callback && callback(null, data);
    }
}

KiKiApi.prototype.submitLetter = function(letter, callback) {
    var self = this;
    var toast_msg = "Posting to KiKi";
    $.ajax({
        url: "/letter",
        method: "POST",
        data: JSON.stringify(letter),
        contentType: "application/json",
        error: self.onError(toast_msg, callback),
        success: self.onSuccess(toast_msg, callback)
    });
}

KiKiApi.prototype.likePost = function(post_id) {
    this.submitLetter({
        "purpose": "action-like",
        "to": ["public"],
        "content": post_id,
    });
}

KiKiApi.prototype.changeName = function(name, callback) {
    this.submitLetter({
        "purpose": "action-assign/name",
        "to": ["public"],
        "content": name
    }, callback);
}

KiKiApi.prototype.followUser = function(user_id) {
    this.submitLetter({
        "purpose": "action-follow",
        "to": ["public"],
        "content": user_id,
    });
}

KiKiApi.prototype.fetchPosts = function(callback) {
    var self = this;
    $.ajax({
        url: "/api/v1/posts",
        method: "GET",
        error: self.onError(null, callback),
        success: self.onSuccess(null, callback)
    });
}

KiKiApi.prototype.fetchPost = function(post_id, callback) {
    var self = this;
    $.ajax({
        url: "/api/v1/post/"+post_id,
        method: "GET",
        error: self.onError(null, callback),
        success: self.onSuccess(null, callback)
    });
}

KiKiApi.prototype.fetchPostComments = function(post_id, callback) {
    var self = this;
    $.ajax({
        url: "/api/v1/post/"+post_id+"/comments",
        method: "GET",
        error: self.onError(null, callback),
        success: self.onSuccess(null, callback)
    });
}

KiKiApi.prototype.fetchUser = function(user_id, callback) {
    var self = this;
    var url = "/api/v1/user"
    if (user_id) {
        url += "/"+user_id
    }
    $.ajax({
        url: url,
        method: "GET",
        error: self.onError(null, callback),
        success: self.onSuccess(null, callback)
    });
}
