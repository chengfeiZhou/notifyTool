define(function () {
    const CHANNEL_SPLITER = '|';

    function Notify(options) {
        options = options || {};
        if(options.autoReconnect === undefined){
            options.autoReconnect = true;
        }
        this.init(options);
        this.connect();
    }

    Notify.prototype.init = function(options) {
        this.address = options.notify_address || (window.location.hostname + ':8900');
        this.wsUrl = 'ws://' + this.address +  '/';
        this.ws = null;
        this.timer = null;
        this.options = options;
        this.channels = {};
        return this;
    };

    Notify.prototype.connect = function() {
        this.ws = new WebSocket(this.wsUrl);
        this.ws.onmessage = this.onmessage.bind(this);
        this.ws.onopen = this.onopen.bind(this);
        this.ws.onclose = this.onclose.bind(this);
        this.ws.onerror = this.onerror.bind(this);
        return this;
    };

    Notify.prototype.onopen = function() {
        this.updateChannelList();
    };

    Notify.prototype.onmessage = function(e) {
        var result = JSON.parse(e.data);
        // console.log(result);
        if (!result.channel) return this;
        var targetChannel = result.channel.split(CHANNEL_SPLITER);
        if (!targetChannel.length) return this;
        targetChannel.map((channel) => {
            if (!!this.channels[channel]) {
                this.channels[channel].map((cb) => {
                    cb(result.content);
                });
            }
        });
        return this;
    };

    Notify.prototype.subscribe = function(channel, cb) {
        if (!this.channels.hasOwnProperty(channel)) {
            this.channels[channel] = [];
            this.updateChannelList();
        }
        this.channels[channel].push(cb);
        return this;
    };

    Notify.prototype.unsubscribe = function(channel,cb) {
        if (!this.channels.hasOwnProperty(channel)) return this;
        if (!cb) {
            delete this.channels[channel];
            this.updateChannelList();
        } else {
            var index = this.channels[channel].indexOf(cb);
            index > -1 && this.channels[channel].splice(index,1);
            if (this.channels[channel].length === 0) {
                delete this.channels[channel];
                this.updateChannelList();
            }
        }
        return this;
    };

    Notify.prototype.updateChannelList = function() {
        (!!this.ws && this.ws.readyState == this.ws.OPEN) && this.ws.send(JSON.stringify({
            command_type:'update_channel',
            message: Object.keys(this.channels).join('|')
        }));
    };

    Notify.prototype.onclose = function(e) {
        clearTimeout(this.timer);
        if (this.options.autoReconnect && e.code != 1000) {
            delete this.ws;
            this.timer = setTimeout(() => {
                this.connect();
            }, 2000);
        }
    };

    Notify.prototype.onerror = function(e) {

    };

    Notify.prototype.close = function() {
        this.ws && this.ws.close(1000);
        this.ws = null;
        clearTimeout(this.timer);
    };
    return Notify;
});