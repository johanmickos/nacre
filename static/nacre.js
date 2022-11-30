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
        status.classList.remove('disconncted');
        status.classList.add('connected');
        socket.send(feedId);
    };
    socket.onclose = function() {
        status.classList.remove('connected');
        status.classList.add('disconnected');
        terminal.options.cursorBlink = false;
    };
}());