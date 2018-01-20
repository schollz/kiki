

var User = function(data, api) {
    this.data = data;
    this.hash;
    this.api = api;
    this.callbacks = [];
}

User.prototype.registerUpdateCallback = function(callback) {
    if ("function" == typeof(callback)) {
        this.callbacks.push(callback);
    }
}

User.prototype.onUpdate = function() {
    for (var i=0; i<this.callbacks.length; i++) {
        this.callbacks[i]();
    }
}

User.prototype.getPublicKey = function() {
    return this.data.public_key;
}

User.prototype.getUsername = function() {
    return this.data.name;
}

User.prototype.getDisplayName = function() {
    return this.data.name || this.data.public_key;
}

User.prototype.getProfile = function() {
    return this.data.profile;
}

User.prototype.getImage = function() {
    return this.data.image;
}

User.prototype.update = function() {
    var self = this;
    if (this.api) {
        this.api.fetchUser(this.getPublicKey(), function(err, res){
            if (err) throw err;
            self.data = res.data.user
            self.onUpdate();
        });
    }
}


var UsersCollection = function(api) {
    this.api = api;
    this.data = {};
    this.callbacks = {};
}

UsersCollection.prototype.fetchUser = function(user_id, callback) {
    var self = this;
    if (this.data[user_id]) {
        return callback(null, this.data[user_id]);
    }
    if (!this.callbacks[user_id]) {
        this.callbacks[user_id] = [];
    }
    this.callbacks[user_id].push(callback);
    // only fire one request
    if (1 < this.callbacks[user_id].length){
        return
    };
    // fetch from api
    this.api.fetchUser(user_id, function(err, res){
        if (err) {
            self.runCallbacks(err, user_id);
            return;
        }
        self.data[user_id] = new User(res.data.user, self.api);
        // self.data[user_id].registerUpdateCallback(callback);
        self.runCallbacks(null, user_id);
    });

}

UsersCollection.prototype.runCallbacks = function(err, user_id) {
    while (this.callbacks[user_id].length) {
        var callback = this.callbacks[user_id].shift();
        callback && callback(err, this.data[user_id]);
    }
}
