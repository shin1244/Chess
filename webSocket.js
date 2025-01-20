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
                }
            }
        }

        if (message.type === 'click') {
            const piece = message.piece;
            const row = message.position.row;
            const col = message.position.col;
            const possibleMoves = calculatePossibleMoves(piece, row, col);
            drawPossibleMoves(possibleMoves);
        }

        // if (message.type === 'move') {
        //     context.drawImage(pieces[message.piece], message.position.col * squareSize, message.position.row * squareSize, squareSize, squareSize);
        // }
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

    function calculatePossibleMoves(piece, row, col) {
        const moves = [];
        const directions = {
            Knight: [
                { row: -2, col: -1 }, { row: -2, col: 1 },
                { row: -1, col: -2 }, { row: -1, col: 2 },
                { row: 1, col: -2 }, { row: 1, col: 2 },
                { row: 2, col: -1 }, { row: 2, col: 1 }
            ],
            Bishop: [
                { row: -1, col: -1 }, { row: -1, col: 1 },
                { row: 1, col: -1 }, { row: 1, col: 1 }
            ],
            Rook: [
                { row: -1, col: 0 }, { row: 1, col: 0 },
                { row: 0, col: -1 }, { row: 0, col: 1 }
            ],
            King: [
                { row: -1, col: -1 }, { row: -1, col: 0 }, { row: -1, col: 1 },
                { row: 0, col: -1 }, { row: 0, col: 1 },
                { row: 1, col: -1 }, { row: 1, col: 0 }, { row: 1, col: 1 }
            ]
        };

        // Pawn 이동 가능성 계산
        if (piece.includes('Pawn')) {
            const direction = piece.includes('white') ? -1 : 1;
            const newRow = row + direction;
            if (newRow >= 0 && newRow < 8) {
                moves.push({ row: newRow, col: col });
            }
        }

        // Knight 이동 가능성 계산
        if (piece.includes('Knight')) {
            directions.Knight.forEach(dir => {
                const newRow = row + dir.row;
                const newCol = col + dir.col;
                if (newRow >= 0 && newRow < 8 && newCol >= 0 && newCol < 8) {
                    moves.push({ row: newRow, col: newCol });
                }
            });
        }

        // Bishop 이동 가능성 계산
        if (piece.includes('Bishop')) {
            directions.Bishop.forEach(dir => {
                for (let i = 1; i < 8; i++) {
                    const newRow = row + dir.row * i;
                    const newCol = col + dir.col * i;
                    if (newRow >= 0 && newRow < 8 && newCol >= 0 && newCol < 8) {
                        moves.push({ row: newRow, col: newCol });
                    } else {
                        break;
                    }
                }
            });
        }

        // Rook 이동 가능성 계산
        if (piece.includes('Rook')) {
            directions.Rook.forEach(dir => {
                for (let i = 1; i < 8; i++) {
                    const newRow = row + dir.row * i;
                    const newCol = col + dir.col * i;
                    if (newRow >= 0 && newRow < 8 && newCol >= 0 && newCol < 8) {
                        moves.push({ row: newRow, col: newCol });
                    } else {
                        break;
                    }
                }
            });
        }

        // King 이동 가능성 계산
        if (piece.includes('King')) {
            directions.King.forEach(dir => {
                const newRow = row + dir.row;
                const newCol = col + dir.col;
                if (newRow >= 0 && newRow < 8 && newCol >= 0 && newCol < 8) {
                    moves.push({ row: newRow, col: newCol });
                }
            });
        }

        return moves;
    }

    function drawPossibleMoves(moves) {
        moves.forEach(move => {
            context.beginPath();
            context.arc((move.col + 0.5) * squareSize, (move.row + 0.5) * squareSize, 5, 0, 2 * Math.PI);
            context.fillStyle = 'red';
            context.fill();
        });
    }
});