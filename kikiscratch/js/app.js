// var server = "http://localhost:8003";
// "/letter";
// "/assign";
// "/open";

// Backbone notes
// http://adrianmejia.com/blog/2012/09/13/backbone-js-for-absolute-beginners-getting-started-part-2/

var messages = [
    {
        message_id: 'uuid4',
        user_id: '_Mr.Prez_',
        username: 'Barack Obama',
        message: 'Have you ever seen a bunny sit near a flower? Well, here you go.\n<img class="materialboxed responsive-img initialized" src="https://bubblesandbeebots.files.wordpress.com/2017/06/bunnyrabbit-large_trans_nvbqzqnjv4bqkm3ycdi1zvq0mt8cxo2c41vse9jsn00kzbur3ixhago.jpg"/>\nThis is indeed a bunny, sitting near a flower.',
        created_at: 1514600002,
        tags: 'rabbit animal',
        replies: [
            {
                message_id: 'another_uuid4',
                user_id: 'zack_attack',
                username: 'zack',
                message: 'Oh, wow!',
                created_at: 1514696602,
                tags: '',
                replies: [
                    {
                        message_id: 'yet_another_uuid4',
                        user_id: 'Stefan is a God King',
                        username: 'StefanRocksMySocks',
                        message: 'I love bunnies!',
                        created_at: 1514796602,
                        tags: '',
                        replies: []
                    }
                ]
            },
            {
                message_id: 'another_uuid4_',
                user_id: 'tedNugent',
                username: 'Ted Nugent',
                message: 'America freedom and stuffs!',
                created_at: 1514699602,
                tags: '',
                replies: []
            }
        ]
    },
    {
        message_id: 'uuid4',
        user_id: 'zack_attack',
        username: 'zack',
        message: 'This is the coolest app.',
        created_at: 1514000002,
        replies: [
            {
                message_id: 'yet_another_uuid4',
                user_id: 'Stefan is a God King',
                username: 'StefanRocksMySocks',
                message: 'Fo shizzle',
                created_at: 1514103382,
                tags: '',
                replies: [

                    {
                        message_id: 'yet_another_uuid4',
                        user_id: 'Stefan is a God King',
                        username: 'StefanRocksMySocks',
                        message: 'Fo shizzle',
                        created_at: 1514103382,
                        tags: '',
                        replies: []
                    }


                ]
            }
        ]
    },
    {
        message_id: 'uuid4',
        user_id: 'system',
        username: 'system',
        message: 'Welcome!',
        created_at: 1500000002,
        replies: []
    }
]


var app = {
    // https://github.com/d3/d3-time-format/blob/master/README.md#timeFormat
    datetimeFormatter: d3.timeFormat('%I:%M%p %B %e, %Y'),
    userColors: {},
    // https://github.com/d3/d3-scale/blob/master/README.md#schemeCategory20c
    // https://github.com/d3/d3-scale-chromatic
    // colorScale: d3.scaleOrdinal(d3.schemeCategory10),
    // colorScale: d3.scaleOrdinal(d3.schemeCategory20),
    // colorScale: d3.scaleOrdinal(d3.schemeCategory20b),
    // colorScale: d3.scaleOrdinal(d3.schemeCategory20c),
    colorScale: d3.scaleOrdinal(d3.schemeSet1),
    // creates and stores colors for every user_id
    getUserColor: function(user_id) {
        if (this.userColors[user_id]) {
            return this.userColors[user_id];
        }
        var n = Object.keys(this.userColors).length;
        var color = this.colorScale(n);
        this.userColors[user_id] = color;
        this.setSettingsToLocalStorage('colors', this.userColors);
        return color;
    },
    getDateTimeDisplay: function(unix_ts) {
        return $('<span>').addClass('datetime right').append(
            // $('<i>').addClass('material-icons').append(
            //     'access_time'
            // ),
            // $('<i>').addClass('fa fa-clock-o fa-1'),
            app.datetimeFormatter(new Date(unix_ts*1000))
        );
    },
    getUserNameDisplay: function(username) {
        return $('<span>').append(
            $('<i>').addClass('material-icons').append('face'),
            '&nbsp;' + username //,
            // $('<i>').addClass('material-icons right').append('person_add'),
            // $('<i>').addClass('material-icons right').append('block')
        );
    },
    getMessageNavBarDisplay: function(data) {
        return $('<nav>').addClass('nav-wrapper message-navbar').css({backgroundColor: app.getUserColor(data.user_id)}).append(
            $('<div>').addClass('col s12').append(
                $('<a>').addClass('breadcrumb white-text').append(
                    app.getUserNameDisplay(data.username).on('click', function(event){
                        event.stopPropagation();
                        console.log('Open modal for follow or block user');
                    }),
                    app.getDateTimeDisplay(data.created_at)
                )
            )
        );
    },
    getMessageContentsDisplay: function(contents) {
        return $('<div>').addClass('message').append(
            (function(){
                var parts = [];
                var chunks = contents.split('\n');
                for (var i=0; i<chunks.length; i++) {
                    parts.push(
                        $('<p>').append(chunks[i])
                    );
                }
                return parts;
            })()
        );
    },

    getMessageToolBarDisplay: function(data) {
        return $('<div>').addClass('message-toolbar').append(
            // $('<span>').addClass('valign-wrapper right').append(
            //     $('<i>').addClass("material-icons right").append('favorite'),
            //     'x' + (data.replies.length || 0)
            // ),
                    $('<div>').append(
                        $('<a>', {title: 'Reply'})
                            .addClass("waves-effect waves-light btn btn-small right").append(
                                $('<i>').addClass('material-icons').append('reply')
                            ).on('click', function(event) {
                                event.stopPropagation();
                                console.log('reply:', data.message_id);
                            }),
                        $('<a>', {title: 'Edit'})
                            .addClass("waves-effect waves-light btn btn-small right").append(
                                $('<i>').addClass('material-icons').append('edit')
                            ).on('click', function(event) {
                                event.stopPropagation();
                                console.log('edit:', data.message_id);
                            })
                    ),
                    $('<br>')
                );
    },




    getMessageReplyDisplay: function(data) {
        return $('<div>').addClass('replies').append(
        // message container
            (function(){
                var replies = [];
                for (var i=0; i<data.replies.length; i++) {
                    var reply = data.replies[i];
                    replies.push(
                        $('<div>').addClass('card').append(
                            app.getMessageNavBarDisplay(reply)
                                .on('click', function(event) {
                                    event.stopPropagation();
                                    // $($(this).find('.card-content')[0]).toggle();
                                    $($(this).parent().find('.card-content')[0]).toggle();
                                }),

                            $('<div>').addClass('card-content message-content').append(
                                $('<div>').addClass('row').append(
                                    $('<div>').addClass('col s10').append(
                                        app.getMessageContentsDisplay(reply.message),
                                    ),
                                    $('<div>').addClass('col s2').append(
                                        app.getMessageToolBarDisplay(reply)
                                    )
                                ),
                                // app.getMessageContentsDisplay(reply.message),
                                // app.getMessageToolBarDisplay(reply),
                                app.getMessageReplyDisplay(reply)

                            ).hide()
                        )
                        // .on('click', function(event) {
                        //     event.stopPropagation();
                        //     $($(this).find('.card-content')[0]).toggle();
                        // })
                    );
                }
                return replies;
            })()
        );
    },

    setSettingsToLocalStorage: function(key, data) {
        localStorage.setItem(key, JSON.stringify(data));
    },

    getSettingsFromLocalStorage: function(key) {
        var data = localStorage.getItem(key);
        if (!data) {
            data = {}
        } else {
            data = JSON.parse(data);
        }
        return data;
    },

    init: function() {
        this.userColors = this.getSettingsFromLocalStorage('colors');
    }
};

app.Message = Backbone.Model.extend({
    defaults: {
        title: '',
        user_id: '',
        message: '',
        created_at: new Date(),
        replies: []
    }
});

app.MessageView = Backbone.View.extend({
    tagName: 'div',

    template: function(data) {
        return $('<div>').addClass('row').append(
                    $('<div>').addClass('col s12 m12').append(
                        $('<div>').addClass('card ').append(
                            app.getMessageNavBarDisplay(data),
                            $('<div>').addClass('card-content message-content').append(
                                $('<div>').addClass('row').append(
                                    $('<div>').addClass('col s10').append(
                                        app.getMessageContentsDisplay(data.message),
                                    ),
                                    $('<div>').addClass('col s2').append(
                                        app.getMessageToolBarDisplay(data)
                                    )
                                ),
                                // app.getMessageContentsDisplay(data.message),
                                // app.getMessageToolBarDisplay(data),
                                app.getMessageReplyDisplay(data)
                            )
                        )
                    )
                );
    },
    render: function(){
        this.$el.html(this.template(this.model.toJSON()));
        return this; // enable chained calls
    }
});


/**
 * Collection of messages
 */
app.Messages = Backbone.Collection.extend({
    model: app.Message
});
// instance of the Collection
app.messages = new app.Messages();



app.AppView = Backbone.View.extend({
    // el - stands for element. Every view has a element associate in with HTML
    //      content will be rendered.
    el: '#container',
    // template which has the placeholder 'who' to be substitute later
    // template: _.template("<h3 id='TEST'>Hello <%= who %></h3>"),
    // It's the first function called when this view it's instantiated.
    initialize: function(){
        // this.render();
        app.messages.on('add', this.addOne, this);
        app.messages.on('reset', this.addAll, this);
    },

    // events: {
    //   'click .card nav': 'toggleMessage'
    // },
    //
    // toggleMessage: function(e) {
    //     var elem = $(e.target).find('.card-content');
    //     console.log(elem);
    //     debugger;
    // },

    addOne: function(message) {
        var view = new app.MessageView({model: message});
        $('.messsages-container').append(view.render().el);
    },

    addAll: function(){
        $('.messsages-container').html(''); // clean the todo list
        app.messages.each(this.addOne, this);
    },

    // $el - it's a cached jQuery object (el), in which you can use jQuery functions
    //       to push content. Like the Hello World in this case.
    render: function(){
        // this.$el.html(this.template({who: 'world!'}));
    }
});


function initApp() {
    app.init();
    app.appView = new app.AppView();
    for (var i=0; i<messages.length; i++) {
        var msg = new app.Message(messages[i]);
        app.messages.add(msg);
    }
}
