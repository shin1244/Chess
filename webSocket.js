const ws = new WebSocket('ws://localhost:3000/ws');
let count = 0;
const colors = ['#00cc00', '#cc8400', '#f0d9b5', '#b58863'];
const tileSize = 80;
const canvas = document.getElementById('chessBoard');
const context = canvas.getContext('2d');

ws.onopen = () => {
    console.log('서버에 연결되었습니다');
};

ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    console.log(data);
    if (data.type === 'count') {
        count = data.count;

        const display = document.getElementById('colorDisplay');
        display.textContent = count;
    }

    if (data.type === 'newGame') {
        for (let row = 0; row < 8; row++) {
            for (let col = 0; col < 8; col++) {
                const x = col * tileSize;
                const y = row * tileSize;
                const colorIndex = (row + col) % 2;
                context.fillStyle = colors[colorIndex+2];
                context.fillRect(x, y, tileSize, tileSize);
            }
        }
    }  
};

canvas.addEventListener('click', (event) => {
    const rect = canvas.getBoundingClientRect();
    const x = event.clientX - rect.left;
    const y = event.clientY - rect.top;
    const col = Math.floor(x / tileSize);
    const row = Math.floor(y / tileSize);

    // 클릭된 타일의 위치를 서버에 전송
    const message = JSON.stringify({ type: 1, row: row, col: col });
    console.log(message);
    ws.send(message);
});

ws.onerror = (error) => {
    document.getElementById('colorDisplay').textContent = '연결 오류!';
};

ws.onclose = () => {
    document.getElementById('colorDisplay').textContent = '연결이 끊어졌습니다';
};