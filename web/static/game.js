// Game state
let ws = null;
let gameState = {
    connected: false,
    roomId: null,
    playerId: null,
    playerName: null,
    isHost: false,
    currentGameState: null,
    myCards: []
};

// DOM Elements
const landingPage = document.getElementById('landing-page');
const gameRoom = document.getElementById('game-room');
const loadingOverlay = document.getElementById('loading-overlay');
const errorModal = document.getElementById('error-modal');
const errorMessage = document.getElementById('error-message');

// Landing page elements
const playerNameInput = document.getElementById('player-name');
const createRoomBtn = document.getElementById('create-room-btn');
const joinRoomBtn = document.getElementById('join-room-btn');
const roomCodeInput = document.getElementById('room-code');

// Game room elements
const currentRoomCode = document.getElementById('current-room-code');
const playersCount = document.getElementById('players-count');
const leaveRoomBtn = document.getElementById('leave-room-btn');
const potAmount = document.getElementById('pot-amount');
const currentBet = document.getElementById('current-bet');
const toCall = document.getElementById('to-call');
const actionPanel = document.getElementById('action-panel');
const hostControls = document.getElementById('host-controls');
const startGameBtn = document.getElementById('start-game-btn');

// Chat elements
const chatMessages = document.getElementById('chat-messages');
const chatInput = document.getElementById('chat-input');
const sendChatBtn = document.getElementById('send-chat-btn');
const gameLog = document.getElementById('game-log');

// Action buttons
const foldBtn = document.getElementById('fold-btn');
const checkBtn = document.getElementById('check-btn');
const callBtn = document.getElementById('call-btn');
const betBtn = document.getElementById('bet-btn');
const raiseBtn = document.getElementById('raise-btn');
const allinBtn = document.getElementById('allin-btn');

// Bet slider
const betSlider = document.getElementById('bet-slider');
const betAmountSlider = document.getElementById('bet-amount-slider');
const betAmountInput = document.getElementById('bet-amount-input');
const confirmBetBtn = document.getElementById('confirm-bet-btn');
const cancelBetBtn = document.getElementById('cancel-bet-btn');

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    setupEventListeners();
    
    // Load saved player name
    const savedName = localStorage.getItem('playerName');
    if (savedName) {
        playerNameInput.value = savedName;
    }
});

// Event Listeners
function setupEventListeners() {
    // Landing page
    createRoomBtn.addEventListener('click', createRoom);
    joinRoomBtn.addEventListener('click', joinRoom);
    roomCodeInput.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') joinRoom();
    });
    
    // Game room
    leaveRoomBtn.addEventListener('click', leaveRoom);
    startGameBtn.addEventListener('click', startGame);
    
    // Chat
    sendChatBtn.addEventListener('click', sendChat);
    chatInput.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') sendChat();
    });
    
    // Tabs
    document.querySelectorAll('.tab').forEach(tab => {
        tab.addEventListener('click', () => switchTab(tab.dataset.tab));
    });
    
    // Action buttons
    foldBtn.addEventListener('click', () => sendAction('fold'));
    checkBtn.addEventListener('click', () => sendAction('check'));
    callBtn.addEventListener('click', () => sendAction('call'));
    betBtn.addEventListener('click', () => showBetSlider('bet'));
    raiseBtn.addEventListener('click', () => showBetSlider('raise'));
    allinBtn.addEventListener('click', () => sendAction('allin'));
    
    // Bet slider
    betAmountSlider.addEventListener('input', updateBetAmount);
    betAmountInput.addEventListener('input', updateBetSlider);
    confirmBetBtn.addEventListener('click', confirmBet);
    cancelBetBtn.addEventListener('click', hideBetSlider);
}

// WebSocket Connection
function connectWebSocket() {
    return new Promise((resolve, reject) => {
        showLoading('Connecting to server...');
        
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        ws = new WebSocket(`${protocol}//${window.location.host}/ws`);
        
        ws.onopen = () => {
            console.log('WebSocket connected');
            gameState.connected = true;
            hideLoading();
            resolve();
        };
        
        ws.onmessage = handleMessage;
        
        ws.onerror = (error) => {
            console.error('WebSocket error:', error);
            hideLoading();
            reject(error);
        };
        
        ws.onclose = () => {
            console.log('WebSocket disconnected');
            gameState.connected = false;
            if (gameState.roomId) {
                showError('Connection lost. Please refresh the page.');
            }
        };
    });
}

// Message Handlers
function handleMessage(event) {
    const message = JSON.parse(event.data);
    console.log('Received:', message);
    
    switch (message.type) {
        case 'welcome':
            handleWelcome(message.data);
            break;
        case 'joinedRoom':
            handleJoinedRoom(message.data);
            break;
        case 'playerJoined':
            handlePlayerJoined(message.data);
            break;
        case 'playerLeft':
            handlePlayerLeft(message.data);
            break;
        case 'gameUpdate':
            handleGameUpdate(message.data);
            break;
        case 'chat':
            handleChatMessage(message.data);
            break;
        case 'error':
            showError(message.data.error);
            break;
    }
}

function handleWelcome(data) {
    console.log('Connected with client ID:', data.clientId);
}

function handleJoinedRoom(data) {
    gameState.roomId = data.roomId;
    gameState.playerId = data.playerId;
    
    // Update UI
    showGameRoom();
    currentRoomCode.textContent = data.room.code;
    
    // Update room state
    updateRoomState(data.room);
    
    // Show host controls if needed
    if (data.room.hostId === gameState.playerId) {
        gameState.isHost = true;
        hostControls.style.display = 'block';
    }
    
    addGameLogEntry(`You joined room ${data.room.code}`);
}

function handlePlayerJoined(data) {
    addGameLogEntry(`${data.player.name} joined the room`);
    // Room state will be updated in the next game update
}

function handlePlayerLeft(data) {
    addGameLogEntry(`Player left the room`);
}

function handleGameUpdate(data) {
    gameState.currentGameState = data.gameState;
    updateGameState(data.gameState);
    
    if (data.action) {
        // Log the action
        const player = data.gameState.players.find(p => p.id === data.playerId);
        if (player) {
            addGameLogEntry(`${player.name} ${data.action}`);
        }
    }
}

function handleChatMessage(data) {
    addChatMessage(data.playerName, data.text);
}

// Room Management
async function createRoom() {
    const playerName = playerNameInput.value.trim();
    if (!playerName) {
        showError('Please enter your name');
        return;
    }
    
    try {
        await connectWebSocket();
        
        // First create room via REST API
        const response = await fetch('/api/rooms/create', { method: 'POST' });
        const data = await response.json();
        
        if (!response.ok) {
            throw new Error(data.error || 'Failed to create room');
        }
        
        // Then join via WebSocket
        localStorage.setItem('playerName', playerName);
        gameState.playerName = playerName;
        
        ws.send(JSON.stringify({
            type: 'joinRoom',
            data: {
                roomId: data.roomId,
                playerName: playerName
            }
        }));
        
    } catch (error) {
        showError(error.message);
    }
}

async function joinRoom() {
    const playerName = playerNameInput.value.trim();
    const roomCode = roomCodeInput.value.trim().toUpperCase();
    
    if (!playerName) {
        showError('Please enter your name');
        return;
    }
    
    if (!roomCode) {
        showError('Please enter room code');
        return;
    }
    
    try {
        await connectWebSocket();
        
        localStorage.setItem('playerName', playerName);
        gameState.playerName = playerName;
        
        ws.send(JSON.stringify({
            type: 'joinRoom',
            data: {
                roomId: roomCode,
                playerName: playerName
            }
        }));
        
    } catch (error) {
        showError(error.message);
    }
}

function leaveRoom() {
    if (confirm('Are you sure you want to leave the room?')) {
        ws.send(JSON.stringify({
            type: 'leaveRoom',
            data: { roomId: gameState.roomId }
        }));
        
        // Reset state and go back to landing
        gameState.roomId = null;
        gameState.playerId = null;
        gameState.isHost = false;
        ws.close();
        
        showLandingPage();
    }
}

function startGame() {
    ws.send(JSON.stringify({
        type: 'startGame',
        data: { roomId: gameState.roomId }
    }));
}

// Game Actions
function sendAction(action, amount = 0) {
    ws.send(JSON.stringify({
        type: 'gameAction',
        data: {
            action: action,
            amount: amount
        }
    }));
    
    hideBetSlider();
}

function showBetSlider(actionType) {
    betSlider.style.display = 'block';
    betSlider.dataset.actionType = actionType;
    
    // Set slider limits based on game state
    const playerState = gameState.currentGameState.players.find(p => p.id === gameState.playerId);
    if (playerState) {
        const minBet = actionType === 'bet' ? 
            gameState.currentGameState.bigBlind : 
            gameState.currentGameState.currentBet + gameState.currentGameState.minRaise;
        
        betAmountSlider.min = minBet;
        betAmountSlider.max = playerState.chips;
        betAmountSlider.value = minBet;
        betAmountInput.value = minBet;
    }
}

function hideBetSlider() {
    betSlider.style.display = 'none';
}

function updateBetAmount() {
    betAmountInput.value = betAmountSlider.value;
}

function updateBetSlider() {
    betAmountSlider.value = betAmountInput.value;
}

function confirmBet() {
    const amount = parseInt(betAmountInput.value);
    const actionType = betSlider.dataset.actionType;
    sendAction(actionType, amount);
}

// Chat
function sendChat() {
    const text = chatInput.value.trim();
    if (!text) return;
    
    ws.send(JSON.stringify({
        type: 'chat',
        data: {
            text: text,
            playerName: gameState.playerName
        }
    }));
    
    chatInput.value = '';
}

function addChatMessage(playerName, text) {
    const messageEl = document.createElement('div');
    messageEl.className = 'chat-message';
    messageEl.innerHTML = `<span class="player-name">${playerName}:</span> ${text}`;
    chatMessages.appendChild(messageEl);
    chatMessages.scrollTop = chatMessages.scrollHeight;
}

function addGameLogEntry(text) {
    const entry = document.createElement('div');
    entry.className = 'log-entry';
    entry.textContent = `[${new Date().toLocaleTimeString()}] ${text}`;
    gameLog.appendChild(entry);
    gameLog.scrollTop = gameLog.scrollHeight;
}

// UI Updates
function updateRoomState(room) {
    playersCount.textContent = `${room.players.length}/${room.maxPlayers}`;
    
    // Update player seats
    for (let i = 0; i < 6; i++) {
        const seat = document.getElementById(`seat-${i}`);
        const player = room.players.find(p => p.seatPosition === i);
        
        if (player) {
            updatePlayerSeat(seat, player);
        } else {
            clearPlayerSeat(seat);
        }
    }
    
    // Update start button
    if (gameState.isHost && room.status === 'ready') {
        startGameBtn.disabled = false;
        startGameBtn.textContent = 'Start Game';
    } else if (room.status === 'waiting') {
        startGameBtn.disabled = true;
        startGameBtn.textContent = `Need ${room.minPlayers - room.players.length} more players`;
    }
}

function updateGameState(state) {
    // Update pot
    potAmount.textContent = state.pot;
    
    // Update community cards
    updateCommunityCards(state.communityCards);
    
    // Update players
    state.players.forEach(player => {
        const seat = document.getElementById(`seat-${player.seatPosition}`);
        if (seat) {
            updatePlayerSeat(seat, player, state);
        }
    });
    
    // Update dealer button
    document.querySelectorAll('.dealer-button').forEach(btn => btn.style.display = 'none');
    const dealerSeat = document.querySelector(`#seat-${state.dealerIndex} .dealer-button`);
    if (dealerSeat) dealerSeat.style.display = 'flex';
    
    // Update action panel
    updateActionPanel(state);
    
    // Handle winners
    if (state.handComplete && state.winners) {
        showWinners(state.winners);
    }
}

function updatePlayerSeat(seat, player, gameState = null) {
    seat.querySelector('.player-name').textContent = player.name;
    seat.querySelector('.player-chips').textContent = `${player.chips} chips`;
    
    if (gameState) {
        // Update bet
        const betEl = seat.querySelector('.player-bet');
        if (player.currentBet > 0) {
            betEl.textContent = player.currentBet;
        } else {
            betEl.textContent = '';
        }
        
        // Update active state
        if (gameState.currentPlayerId === player.id) {
            seat.classList.add('active');
        } else {
            seat.classList.remove('active');
        }
        
        // Update folded state
        if (player.isFolded) {
            seat.classList.add('folded');
        } else {
            seat.classList.remove('folded');
        }
    }
}

function clearPlayerSeat(seat) {
    seat.querySelector('.player-name').textContent = 'Empty Seat';
    seat.querySelector('.player-chips').textContent = '-';
    seat.querySelector('.player-bet').textContent = '';
    seat.classList.remove('active', 'folded');
}

function updateCommunityCards(cards) {
    const slots = ['flop-1', 'flop-2', 'flop-3', 'turn', 'river'];
    
    slots.forEach((slotId, index) => {
        const slot = document.getElementById(slotId);
        if (cards && index < cards.length) {
            const card = cards[index];
            slot.textContent = card.display;
            slot.className = `card-slot suit-${card.suit}`;
        } else {
            slot.textContent = '';
            slot.className = 'card-slot';
        }
    });
}

function updateActionPanel(state) {
    const myPlayer = state.players.find(p => p.id === gameState.playerId);
    if (!myPlayer || myPlayer.isFolded || state.handComplete) {
        actionPanel.style.display = 'none';
        return;
    }
    
    // Show panel only if it's my turn
    if (state.currentPlayerId === gameState.playerId) {
        actionPanel.style.display = 'block';
        
        // Update betting info
        currentBet.textContent = state.currentBet;
        const callAmount = state.currentBet - myPlayer.currentBet;
        toCall.textContent = callAmount;
        
        // Update button states
        checkBtn.style.display = state.currentBet === myPlayer.currentBet ? 'inline-block' : 'none';
        callBtn.style.display = callAmount > 0 ? 'inline-block' : 'none';
        callBtn.querySelector('#call-amount').textContent = callAmount > 0 ? callAmount : '';
        betBtn.style.display = state.currentBet === 0 ? 'inline-block' : 'none';
        raiseBtn.style.display = state.currentBet > 0 ? 'inline-block' : 'none';
        
        // Disable buttons if not enough chips
        if (callAmount > myPlayer.chips) {
            callBtn.disabled = true;
        }
    } else {
        actionPanel.style.display = 'none';
    }
}

function showWinners(winners) {
    winners.forEach(winner => {
        const player = gameState.currentGameState.players.find(p => p.id === winner.playerId);
        if (player) {
            addGameLogEntry(`${player.name} wins ${winner.amount} chips with ${winner.description}`);
        }
    });
}

// UI Navigation
function showGameRoom() {
    landingPage.classList.remove('active');
    gameRoom.classList.add('active');
}

function showLandingPage() {
    gameRoom.classList.remove('active');
    landingPage.classList.add('active');
}

function switchTab(tabName) {
    document.querySelectorAll('.tab').forEach(tab => {
        tab.classList.toggle('active', tab.dataset.tab === tabName);
    });
    
    document.querySelectorAll('.tab-pane').forEach(pane => {
        pane.classList.toggle('active', pane.id === `${tabName}-tab`);
    });
}

// Loading & Error
function showLoading(message = 'Loading...') {
    loadingOverlay.style.display = 'flex';
    loadingOverlay.querySelector('p').textContent = message;
}

function hideLoading() {
    loadingOverlay.style.display = 'none';
}

function showError(message) {
    errorMessage.textContent = message;
    errorModal.style.display = 'flex';
}

function closeErrorModal() {
    errorModal.style.display = 'none';
}