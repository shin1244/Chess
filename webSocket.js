function loadPiece(piecePath) {
    const img = new Image();
    img.src = piecePath;
    return img;
}


$(document).ready(function() {
    const canvas = $('#chessboard')[0];
    const context = canvas.getContext('2d');
    const squareSize = 80;
    const pieces = {
        whitePawn: loadPiece('assets/whitePawn.png'),
        blackPawn: loadPiece('assets/blackPawn.png'),
        whiteKnight: loadPiece('assets/whiteKnight.png'),
        blackKnight: loadPiece('assets/blackKnight.png'),
        whiteBishop: loadPiece('assets/whiteBishop.png'),
        blackBishop: loadPiece('assets/blackBishop.png'),
        whiteRook: loadPiece('assets/whiteRook.png'),
        blackRook: loadPiece('assets/blackRook.png'),
        whiteKing: loadPiece('assets/whiteKing.png'),
        blackKing: loadPiece('assets/blackKing.png'),
    };

    
    for (let row = 0; row < 8; row++) {
        for (let col = 0; col < 8; col++) {
            context.fillStyle = (row + col) % 2 === 0 ? 'white' : 'gray';
            context.fillRect(col * squareSize, row * squareSize, squareSize, squareSize);
        }
    }

    const socket = new WebSocket('ws://localhost:3000/ws');

    socket.onmessage = function(event) {
        const message = JSON.parse(event.data);
        if (message.type === 'spawn') {
            console.log(message.piece);
            console.log(message.position);
            console.log(pieces[message.piece]);
            context.drawImage(pieces[message.piece], message.position.col * squareSize, message.position.row * squareSize, squareSize, squareSize);
        }
    };

    $(canvas).on('click', function(event) {
        const rect = canvas.getBoundingClientRect();
        const x = event.clientX - rect.left;
        const y = event.clientY - rect.top;
        
        // 클릭한 위치의 체스판 좌표 계산
        const col = Math.floor(x / squareSize);
        const row = Math.floor(y / squareSize);
        
        // 웹소켓으로 클릭 위치 전송
        const message = {
            type: 'click',
            position: {
                row: row,
                col: col
            }
        };
        
        socket.send(JSON.stringify(message));
    });
});