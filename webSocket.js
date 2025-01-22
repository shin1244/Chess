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

    const socket = new WebSocket('ws://localhost:3000/ws');
    let color = null; // 플레이어의 색상 정보를 저장할 변수

    socket.onopen = function() {
        console.log("연결 성공");
    };

    socket.onmessage = function(event) {
        const message = JSON.parse(event.data);
        console.log('받은 메시지:', message); // 디버깅용 로그

        if (message.type === 'color') {
            color = message.player_color; // 서버로부터 받은 색상 정보를 저장
            console.log(color);
        }

        if (message.type === 'spawn') {
            context.drawImage(pieces[message.piece], message.position.col * squareSize, message.position.row * squareSize, squareSize, squareSize);
        }

        if (message.type === 'board') {
            const board = message.board;
            for (let row = 0; row < 8; row++) {
                for (let col = 0; col < 8; col++) {
                    context.fillStyle = board[row][col].Color === 2 ? '#6b8e23' : '#d3d3d3';
                    context.fillRect(col * squareSize, row * squareSize, squareSize, squareSize);
                    if (board[row][col].Piece !== "") {
                        context.drawImage(pieces[board[row][col].Piece], col * squareSize, row * squareSize, squareSize, squareSize);
                    }
                    if (message.goal.some(pos => pos.row === row && pos.col === col)) {
                        context.strokeStyle = "red";
                        context.lineWidth = 3;
                        context.strokeRect(col * squareSize, row * squareSize, squareSize, squareSize);
                    }
                }
            }
        }

        if (message.type === 'click') {
            message.position.forEach(position => {
            context.beginPath();
            context.arc(position.col * squareSize + squareSize/2, position.row * squareSize + squareSize/2, 10, 0, 2 * Math.PI);
            context.fillStyle = "red";
            context.fill();
            context.closePath();
            });
        }
    };

    $(canvas).on('click', function(event) {
        const rect = canvas.getBoundingClientRect();
        const x = event.clientX - rect.left;
        const y = event.clientY - rect.top;
        
        // 클릭한 위치의 체스판 좌표 계산
        const col = Math.floor(x / squareSize);
        const row = Math.floor(y / squareSize);
        
        // 웹소켓으로 클릭 위치와 플레이어 색상 전송
        const message = {
            type: 'click',
            player_color: color,
            position: {
                row: row,
                col: col
            },
        };
        
        socket.send(JSON.stringify(message));
    });
});