
var Post = function(data, api, parent) {
    this.data = data;
    this.hash;
    this.api = api;
    this.parent = parent;
    this.comments = null;
    this.owner = null;
    this.callbacks = [];
    this.$el = this.buildUi();
    this.fetchComments();
    this.fetchOwner();
}

Post.prototype.setParent = function(parent) {
    this.parent = parent;
}

Post.prototype.registerUpdateCallback = function(callback) {
    if ("function" == typeof(callback)) {
        this.callbacks.push(callback);
    }
}

Post.prototype.getTimestamp = function() {
    return this.data.timestamp;
}

Post.prototype.getContent = function() {
    return this.data.content;
}

Post.prototype.getPostId = function() {
    return this.data.id;
}

Post.prototype.getOwnerId = function() {
    return this.data.owner_id;
}

Post.prototype.fetchOwner = function(callback) {
    var self = this;
    if (this.owner) {
        callback && callback(null, this.owner);
        return;
    }
    app.Users.fetchUser(this.getOwnerId(), function(err, user){
        self.owner = user;
        callback && callback(err, user);
    });
}

Post.prototype.getRecipients = function() {
    return this.data.recipients;
}

Post.prototype.onUpdate = function() {
    this.elements.content.html(this.getContent());
    this.elements.likebutton.find("span").text("x"+(this.data.likes||0));
    this.fetchComments();
    for (var i=0; i<this.callbacks.length; i++) {
        this.callbacks[i]();
    }
}

Post.prototype.update = function(data) {
    var self = this;
    if (data) {
        this.data = data;
        this.onUpdate();
        return;
    }
    this.api.fetchPost(this.getPostId(), function(err, res){
        if (err) throw err;
        self.data = res.data.posts[0];
        self.onUpdate();
    });
}

Post.prototype.fetchComments = function() {
    var self = this;
    if (0 < this.data.num_comments) {
        app.Posts.fetchPostComments(this.getPostId(), function(err, posts) {
            self.comments = posts;
            self.buildCommentUi();
        });
    } else {
        this.comments = [];
    }
}

Post.prototype.buildCommentUi = function() {
    var self = this;
    this.elements.comments.html('');
    for (var i=0; i<self.comments.length; i++) {
        app.Ui.collectUsersFromPost(self.comments[i]);
        this.elements.comments.append(
            $('<div>').addClass('col-12').append(
                self.comments[i].$el
            )
        );
    }
}

Post.prototype.toLetter = function(action) {
    if ("edit" == action) {
        return {
            replaces: this.getPostId(),
            purpose: 'share-text',
            reply_to: this.data.reply_to,
            content: this.getContent()
        }
    } else if ("reply" == action) {
        return {
            purpose: 'share-text',
            reply_to: this.getPostId()
        }
    } else {
        throw new Error("action not found");
    }
}

Post.prototype.buildUi = function() {
    var self = this;
    this.elements = {
        content: $('<div>').append(
            self.getContent()
        ),
        comments: $('<div>').addClass('post-comments row'),
        datetime: $("<div>").addClass('text-muted post-datetime').append(
            new Date(this.getTimestamp()*1000)
        ),
        likebutton: $("<a>").addClass("post-action likebutton").append(
            $("<i>").addClass("fas fa-heart"),
            $("<span>").text("x"+(self.data.likes||0))
        ).on('click', function() {
            self.api.likePost(self.getPostId(), function(err, res){
                if (err) throw err;
                self.update();
            });
        })
    }
    return $("<div>").addClass('post-container').append(
            $("<div>").addClass("post-header").append(

                $("<span>").addClass("float-right post-controls").append(
                    $("<a>").addClass("post-action").append(
                        $('<i>').addClass("fas fa-reply")
                    ).on('click', function() {
                        app.openModal("Reply to " + self.data.owner_name, self.toLetter("reply"));
                    }),

                    // edit button
                    (function() {
                        // only display if current user owns the post
                        if (self.getOwnerId() == app.User.getPublicKey()) {
                            return $("<a>").addClass("post-action").append(
                                        $("<i>").addClass("fas fa-edit")
                                    ).on('click', function() {
                                        app.openModal("Edit post", self.toLetter("edit"));
                                    });
                        }
                        return null;
                    })(),
                    //.end

                    self.elements.likebutton
                ),

                $("<img>", {
                    owner_id: self.getOwnerId()
                }).addClass("align-middle profile-pic"),

                // activatenamemodal
                $("<a>",{
                    'href':"#!",
                    'data-publickey': self.getOwnerId(),
                }).addClass("activatenamemodal").append(
                     self.data.owner_name || self.getOwnerId()
                ).on('click', function() {
                    self.fetchOwner(function(err, user){
                        if (err) throw err;
                        if (user) {
                            $("#modalNameName").text(user.getDisplayName());

                            $("#modalNamePublicKey").text("");
                            if (user.name) {
                                $("#modalNamePublicKey").text(user.getPublicKey());
                            }

                            $("#modalNameContent").html(user.getProfile());

                            $("#modalNameImage").attr("src", null);
                            if (user.image && "" != user.image) {
                                $("#modalNameImage").attr("src", '/img/'+user.getImage());
                            }

                            $("#nameModal").modal();
                            $("#followButton").attr("data-publickey", user.getPublicKey());
                        }
                    })
                }),
                //.end

                // recipients
                $("<span>").append(
                    " to ",
                    (function(){
                        var elems = [];
                        var recipients = self.getRecipients();
                        for (var i=0; i<recipients.length; i++) {
                            // dont display owner
                            // owner can always see their data
                            if (self.getOwnerId() != recipients[i]) {
                                elems.push(
                                    "[",
                                    $("<span>", {
                                        owner_id: recipients[i]
                                    }).addClass('recipient').append(recipients[i]),
                                    "]"
                                );
                            }
                        }
                        return elems;
                    })()
                ),
                //.end

                self.elements.datetime
            ),

            self.elements.content,
            self.elements.comments
        );
}


var PostsCollection = function(api) {
    this.api = api;
    this.data = {};
    this.callbacks = {};
}

PostsCollection.prototype.addPost = function(data, post_id) {
    var post;
    var parent_post = null;
    if (post_id) {
        parent_post = this.data[post_id];
    }
    if (this.data[data.id]) {
        post = this.data[data.id];
        post.update(data);
        post.setParent(parent_post);
    } else {
        post = new Post(data, this.api, parent_post);
        this.data[data.id] = post;
    }
    return post;
}

PostsCollection.prototype.addPosts = function(data, post_id) {
    var posts = [];
    for (var i=0; i<data.length; i++) {
        posts.push(this.addPost(data[i], post_id));
    }
    return posts;
}

PostsCollection.prototype.fetchPosts = function(callback){
    var self = this;
    this.api.fetchPosts(function(err, res){
        if (err) {
            callback && callback(err, res);
            return
        }
        var posts = [];
        if(!err && res.data.posts) {
            posts = self.addPosts(res.data.posts);
        }
        callback && callback(err, posts);
    });
}

PostsCollection.prototype.fetchPost = function(post_id, callback) {
    var self = this;
    // fetch from api
    this.api.fetchPost(post_id, function(err, res){
        if (err) {
            callback && callback(err);
            return;
        }
        if (res.data.posts && res.data.posts[0]) {
            self.data[post_id] = new Post(res.data.posts[0], self.api);
            callback && callback(null, self.data[post_id]);
        } else {
            callback && callback(new Error("Not found"));
        }
    });
}

PostsCollection.prototype.fetchPostComments = function(post_id, callback) {
    var self = this;
    // fetch from api
    this.api.fetchPostComments(post_id, function(err, res){
        if (err) {
            // self.runCommentCallbacks(err, post_id);
            callback && callback(err);
            return;
        }
        if (res.data.posts) {
            self.data[post_id].comments = self.addPosts(res.data.posts, post_id);
        } else {
            self.data[post_id].comments = [];
        }
        // self.runCommentCallbacks(null, post_id);
        callback && callback(err, self.data[post_id].comments);
    });
}
//
// PostsCollection.prototype.runCommentCallbacks = function(err, post_id) {
//     while (this.callbacks['comments_'+post_id].length) {
//         var callback = this.callbacks['comments_'+post_id].shift();
//         callback && callback(err, this.data[post_id].comments);
//     }
//     if (0 == this.callbacks['comments_'+post_id].length) {
//         delete this.callbacks['comments_'+post_id];
//     }
// }
//
// PostsCollection.prototype.runCallbacks = function(err, post_id) {
//     while (this.callbacks[post_id].length) {
//         var callback = this.callbacks[post_id].shift();
//         callback && callback(err, this.data[post_id]);
//     }
//     if (0 == this.callbacks[post_id].length) {
//         delete this.callbacks[post_id];
//     }
// }
