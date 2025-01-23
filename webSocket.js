function loadPiece(piecePath) {
    const img = new Image();
    img.src = piecePath;
    return img;
}

let socket;
let reconnectAttempts = 0;
const maxReconnectAttempts = 5;
let context;  // 전역 변수로 선언
let canvas;   // 전역 변수로 선언
let squareSize;  // 전역 변수로 선언
let pieces;   // 전역 변수로 선언
let boardColor;  // 전역 변수로 선언
let color;    // 전역 변수로 선언

function connectWebSocket() {
    try {
        // WebSocket URL을 현재 호스트 기준으로 설정
        const wsProtocol = window.location.protocol === 'https:' ? 'wss://' : 'ws://';
        const wsUrl = wsProtocol + window.location.host + '/ws';  // ':30' 부분 제거
        console.log("Connecting to WebSocket:", wsUrl);
        
        socket = new WebSocket(wsUrl);

        socket.onopen = function() {
            console.log("WebSocket 연결됨");
            reconnectAttempts = 0;
        };

        socket.onclose = function() {
            console.log("WebSocket 연결 끊김");
            if (reconnectAttempts < maxReconnectAttempts) {
                console.log("재연결 시도 중...");
                reconnectAttempts++;
                setTimeout(connectWebSocket, 2000);
            }
        };

        socket.onerror = function(error) {
            console.error("WebSocket 에러:", error);
        };

        // 기존 메시지 핸들러 유지
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
                console.log(board);
                for (let row = 0; row < 8; row++) {
                    for (let col = 0; col < 8; col++) {
                        // 타일 배경색 채우기
                        context.fillStyle = boardColor[board[row][col].Color];
                        context.fillRect(col * squareSize, row * squareSize, squareSize, squareSize);
                        
                        // 타일 테두리 그리기
                        context.strokeStyle = '#000000';  // 검은색 테두리
                        context.lineWidth = 1;  // 얇은 선 두께
                        context.strokeRect(col * squareSize, row * squareSize, squareSize, squareSize);
                        
                        // 체스 말 그리기
                        if (board[row][col].Piece !== "") {
                            context.drawImage(pieces[board[row][col].Piece], col * squareSize, row * squareSize, squareSize, squareSize);
                        }
                        
                        // 목표 위치 표시
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

            if (message.type === 'gameOver') {
                let resultMessage = '';
                console.log(message.player_color);
                if (message.player_color === color) {
                    resultMessage = '축하합니다! 승리하셨습니다! 🎉'
                } else {
                    resultMessage = '아쉽네요, 패배하셨습니다. 😢';
                }

                // 모달 팝업 생성 및 표시
                const modalHtml = `
                    <div id="gameOverModal" class="modal">
                        <div class="modal-content">
                            <h2>게임 종료</h2>
                            <p>${resultMessage}</p>
                            <button onclick="location.reload()">새 게임</button>
                        </div>
                    </div>
                `;
                
                $('body').append(modalHtml);
            }
        };
    } catch (error) {
        console.error("WebSocket 연결 실패:", error);
    }
}

$(document).ready(function() {
    canvas = $('#chessboard')[0];
    context = canvas.getContext('2d');
    squareSize = 80;
    pieces = {
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
    boardColor = ['#FF69B4', '#4169E1', '#6b8e23', '#d3d3d3'];
    color = null;

    // 초기 체스판 그리기 추가
    drawInitialBoard();
    
    connectWebSocket();

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

    $('#joinGame').on('click', function() {
        socket.send(JSON.stringify({ type: 'join' }));
    });
});

// 초기 체스판을 그리는 함수 추가
function drawInitialBoard() {
    for (let row = 0; row < 8; row++) {
        for (let col = 0; col < 8; col++) {
            // 체크무늬 패턴으로 타일 그리기
            context.fillStyle = (row + col) % 2 === 0 ? boardColor[2] : boardColor[3];
            context.fillRect(col * squareSize, row * squareSize, squareSize, squareSize);
            
            // 타일 테두리 그리기
            context.strokeStyle = '#000000';
            context.lineWidth = 1;
            context.strokeRect(col * squareSize, row * squareSize, squareSize, squareSize);
        }
    }
}