function loadPiece(piecePath) {
    return new Promise((resolve, reject) => {
        const img = new Image();
        img.onload = () => resolve(img);
        img.onerror = () => reject(new Error(`Failed to load image: ${piecePath}`));
        img.src = piecePath;
    });
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

        socket = new WebSocket(wsUrl);

        socket.onopen = function () {
            console.log("WebSocket 연결됨");
            reconnectAttempts = 0;
        };

        socket.onclose = function () {
            console.log("WebSocket 연결 끊김");
            if (reconnectAttempts < maxReconnectAttempts) {
                console.log("재연결 시도 중...");
                reconnectAttempts++;
                setTimeout(connectWebSocket, 2000);
            }
        };

        socket.onerror = function (error) {
            console.error("WebSocket 에러:", error);
        };

        // 기존 메시지 핸들러 유지
        socket.onmessage = function (event) {
            const message = JSON.parse(event.data);

            if (message.type === 'color') {
                color = message.player_color; // 서버로부터 받은 색상 정보를 저장
                // 게임 로그창 초기화
                $('#logContent').empty();
                $('#joinGame').text('기다리는 중...');
                const playerColor = message.player_color === 0 ? '백' : '흑';
                addLogMessage(`당신은 ${playerColor}입니다.`);
                addLogMessage('기다리는 중...');
            }

            if (message.type === 'board') {
                const board = message.board;

                // 현재 턴 표시를 위한 캔버스 패딩 영역 색상 설정
                canvas.style.backgroundColor = message.turn === color ? '#32CD32' : '#FFFFFF';

                const isBlack = color === 1;
                const transformedBoard = transformBoard(board, isBlack);

                
                for (let row = 0; row < 8; row++) {
                    for (let col = 0; col < 8; col++) {
                        // 타일 배경색 채우기
                        context.fillStyle = boardColor[transformedBoard[row][col].color];
                        context.fillRect(col * squareSize, row * squareSize, squareSize, squareSize);

                        if (message.sound_type === 1) {
                            startSound.play();
                        } else if (message.sound_type === 2) {
                            pieceSound.play();
                        } else if (message.sound_type === 3) {
                            moveSound.play();
                        }

                        // 타일 카운트 표시
                        $('.white-tile-count').text(message.printing_tiles[0]);
                        $('.black-tile-count').text(message.printing_tiles[1]);

                        // 타일 테두리 그리기
                        context.strokeStyle = '#000000';  // 검은색 테두리
                        context.lineWidth = 1;  // 얇은 선 두께
                        context.strokeRect(col * squareSize, row * squareSize, squareSize, squareSize);

                        // 체스 말 그리기
                        if (transformedBoard[row][col].piece !== "") {
                            context.drawImage(pieces[transformedBoard[row][col].piece],
                                col * squareSize, row * squareSize, squareSize, squareSize);
                        }

                        // 목표 위치 표시
                        if (transformedBoard[row][col].goal !== -1) {
                            context.strokeStyle = transformedBoard[row][col].goal === 0 ? "red" : "blue";
                            context.lineWidth = 4;
                            // 테두리를 약간 안쪽으로 그리기
                            context.strokeRect(
                                col * squareSize + 2,
                                row * squareSize + 2,
                                squareSize - 4,
                                squareSize - 4
                            );
                        }
                    }
                }
            }

            if (message.type === 'click') {
                message.positions.forEach(position => {
                    context.beginPath();
                    const transformedPos = transformPosition(position.row, position.col, color === 1, position.piece);
                    context.arc(
                        transformedPos.col * squareSize + squareSize / 2,
                        transformedPos.row * squareSize + squareSize / 2,
                        10, 0, 2 * Math.PI
                    );
                    // 킹의 이동 가능 위치는 파란색, 그 외는 빨간색으로 표시
                    context.fillStyle = transformedPos.piece.includes("King") ? "blue" : "red";
                    context.fill();
                    context.closePath();
                });
            }

            if (message.type === 'gameOver') {
                let resultMessage = '';
                // goals가 2차원 배열이므로 플레이어의 목표 위치에 접근
                if (message.goals && Array.isArray(message.goals)) {
                    message.goals.forEach((playerGoals, playerIndex) => {
                        playerGoals.forEach(goal => {
                            context.strokeStyle = playerIndex === 0 ? "red" : "blue";
                            context.lineWidth = 4;
                            context.strokeRect(
                                goal.col * squareSize + 2,
                                goal.row * squareSize + 2,
                                squareSize - 4,
                                squareSize - 4
                            );
                        });
                    });
                }
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
                            <button onclick="restartGame()">새 게임</button>
                        </div>
                    </div>
                `;

                $('body').append(modalHtml);
                const winner = message.player_color === 0 ? '백' : '흑';
                if (message.piece === "Pawn") {
                    addLogMessage(`게임 종료! ${winner}의 승리입니다! [시간 승리]`);
                } else if (message.piece === "King") {
                    addLogMessage(`게임 종료! ${winner}의 승리입니다! [체크 메이트]`);
                } else if (message.piece === "Rook") {
                    addLogMessage(`게임 종료! ${winner}의 승리입니다! [목표 완료]`);
                }
                gameOverSound.play();
            }
        };
    } catch (error) {
        console.error("WebSocket 연결 실패:", error);
    }
}

$(document).ready(async function () {
    canvas = $('#chessboard')[0];
    context = canvas.getContext('2d');
    squareSize = 80;

    // 캔버스 패딩을 8px로 증가
    canvas.style.padding = '8px';
    canvas.style.boxSizing = 'content-box';
    canvas.style.border = '1px solid #000000';

    // 이미지 로딩을 비동기로 처리
    try {
        moveSound = new Audio('assets/spray.mp3');
        moveSound.volume = 0.5;
        gameOverSound = new Audio('assets/blop.mp3');
        gameOverSound.volume = 0.5;
        startSound = new Audio('assets/board.mp3');
        startSound.volume = 0.5;
        pieceSound = new Audio('assets/piece.mp3');

        pieces = {
            whitePawn: await loadPiece('assets/whitePawn.png'),
            blackPawn: await loadPiece('assets/blackPawn.png'),
            whiteKnight: await loadPiece('assets/whiteKnight.png'),
            blackKnight: await loadPiece('assets/blackKnight.png'),
            whiteBishop: await loadPiece('assets/whiteBishop.png'),
            blackBishop: await loadPiece('assets/blackBishop.png'),
            whiteRook: await loadPiece('assets/whiteRook.png'),
            blackRook: await loadPiece('assets/blackRook.png'),
            whiteKing: await loadPiece('assets/whiteKing.png'),
            blackKing: await loadPiece('assets/blackKing.png'),
        };

        boardColor = ['#FF69B4', '#4169E1', '#6b8e23', '#d3d3d3'];
        color = null;

        // 초기 체스판 그리기
        drawInitialBoard();
        connectWebSocket();

        $(canvas).on('click', function (event) {
            const rect = canvas.getBoundingClientRect();
            const x = event.clientX - rect.left;
            const y = event.clientY - rect.top;

            let col = Math.floor(x / squareSize);
            let row = Math.floor(y / squareSize);

            const transformedPos = transformPosition(row, col, color === 1);

            const message = {
                type: 'click',
                player_color: color,
                position: transformedPos,
                piece: transformedPos.piece
            };

            sendMessage(socket, message);
        });

        $('#joinGame').on('click', function () {
            sendMessage(socket, { type: 'join' });
        });
    } catch (error) {
        console.error('이미지 로딩 실패:', error);
    }
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

function restartGame() {
    $('.modal').remove();
    socket.send(JSON.stringify({ type: 'restart' }));
}

function addLogMessage(message, type = 'system') {
    const logContent = $('#logContent');
    const messageElement = $('<div>')
        .addClass('log-message')
        .addClass(type)
        .text(message);

    logContent.append(messageElement);
    logContent.scrollTop(logContent[0].scrollHeight);
}

function sendChatMessage() {
    const input = $('#chatInput');
    const message = input.val().trim();

    if (message) {
        socket.send(JSON.stringify({
            type: 'chat',
            message: message,
            player_color: color
        }));
        input.val('');
    }
}

// 보드 변환을 위한 유틸리티 함수들
function transformBoard(board, isBlack) {
    if (!isBlack) return board;

    const transformedBoard = JSON.parse(JSON.stringify(board)); // 깊은 복사
    transformedBoard.reverse();
    transformedBoard.forEach(row => row.reverse());
    return transformedBoard;
}

function transformPosition(row, col, isBlack, piece) {
    if (!isBlack) return { row, col, piece };
    return {
        row: 7 - row,
        col: 7 - col,
        piece: piece
    };
}

function sendMessage(socket, message) {
    if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify(message));
    } else {
        console.log("WebSocket이 연결되지 않았거나 닫혔습니다.");
    }
}
