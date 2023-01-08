function subscribe(channel, onmessage, options) {
    require(['common/notify'],function (Notify) {
        if (!notify) {
            notify = new Notify();
        }
        notify.subscribe(channel,onmessage);
    });
    return {
        close: function() {
            notify.unsubscribe(channel,onmessage);
        }
    }
}