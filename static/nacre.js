const CLOSE_TOO_MANY_PEERS = 4001;

(function() {
    const terminal = new Terminal({
        convertEol: true,
        scrollback: 10_000,
        disableStdin: true,
        cursorBlink: true,
    });
    const fitAddon = new FitAddon.FitAddon();
    terminal.loadAddon(fitAddon);
    terminal.open(document.getElementById('terminal'), {focus: true});
    const status = document.getElementById('status');
    fitAddon.fit();
    window.addEventListener('resize', function() {
        fitAddon.fit();
    });
    const url = 'ws://' + window.location.host + '/websocket';
    const socket = new WebSocket(url);
    socket.binaryType = 'arraybuffer';
    const decoder = new TextDecoder('utf-8');
    socket.onmessage = function(ev) {
        terminal.write(decoder.decode(ev.data));
     };
    socket.onopen = function() {
        status.classList.remove('disconncted', 'error');
        status.classList.add('connected');
        socket.send(feedId);
    };
    socket.onclose = function(ev) {
        status.classList.remove('connected', 'error');
        terminal.options.cursorBlink = false;
        switch (ev.code) {
            case CLOSE_TOO_MANY_PEERS:
                status.classList.add('error');
                const state = status.getElementsByClassName('details')[0];
                state.innerHTML = ev.reason;
                break;
            default:
                status.classList.add('disconnected');
        }
    };
    socket.onerror = function() {
        status.classList.remove('connected', 'disconnected');
        status.classList.add('error');
        terminal.options.cursorBlink = false;
    };
}());