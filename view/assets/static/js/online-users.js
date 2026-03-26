document.addEventListener("DOMContentLoaded", () => {
  const sessionID = window.SESSION_ID;
  const currentUserID = Number(window.USER_ID);
  const listEl = document.querySelector(".users-list");
  const chatModal = document.getElementById("chatModal");
  const chatTitle = document.getElementById("chatModalTitle");
  const chatMessages = document.getElementById("chatMessages");
  const chatForm = document.getElementById("chatForm");
  const chatInput = document.getElementById("chatInput");

  if (
    !sessionID ||
    !listEl ||
    !chatModal ||
    !chatTitle ||
    !chatMessages ||
    !chatForm ||
    !chatInput
  )
    return;

  const wsProtocol = location.protocol === "https:" ? "wss" : "ws";
  let ws = null;
  let reconnectTimer = null;
  let reconnectAttempts = 0;
  let isConnecting = false;

  function scheduleReconnect() {
    if (reconnectTimer) return;
    const delay = Math.min(10000, 1000 * Math.pow(2, reconnectAttempts));
    reconnectAttempts += 1;
    reconnectTimer = setTimeout(() => {
      reconnectTimer = null;
      connectWebSocket();
    }, delay);
  }

  function connectWebSocket() {
    if (isConnecting) return;
    if (
      ws &&
      (ws.readyState === WebSocket.OPEN ||
        ws.readyState === WebSocket.CONNECTING)
    ) {
      return;
    }
    isConnecting = true;
    ws = new WebSocket(`${wsProtocol}://${location.host}/ws`);

    ws.onopen = () => {
      console.log("Connecté au WebSocket");
      isConnecting = false;
      reconnectAttempts = 0;
    };

    ws.onmessage = (event) => {
      const data = event.data;

      const parts = data.split("|", 4);
      if (parts.length < 4) return;

      const senderID = Number(parts[0]);
      const receiverID = Number(parts[1]);
      const timestamp = parts[2];
      const content = parts[3];

      if (senderID === 0 && receiverID === 0) {
        if (content === "requestsessionid") {
          if (ws && ws.readyState === WebSocket.OPEN) {
            ws.send(`0|0|${nowString()}|sessionid:${sessionID}`);
          }
          handshakeDone = true;
          sendSystem("sendallactiveusers");
          if (!refreshTimer) {
            refreshTimer = setInterval(() => {
              sendSystem("sendallactiveusers");
            }, 5000);
          }
          return;
        }

        if (content.startsWith("activeusers:")) {
          handleActiveUsers(content);
          return;
        }

        if (content.startsWith("userinfo:")) {
          handleUserInfo(content);
          return;
        }

        if (content.startsWith("unreadusers:")) {
          handleUnreadUsers(content);
          return;
        }

        if (content === "message initial du serveur") {
          historyLoaded = true;
          return;
        }
      }

      if (senderID && receiverID) {
        const otherUserID = senderID === currentUserID ? receiverID : senderID;
        addMessage(
          otherUserID,
          {
            senderID,
            receiverID,
            timestamp,
            content,
          },
          !historyLoaded,
        );
      }
    };

    ws.onclose = () => {
      if (refreshTimer) {
        clearInterval(refreshTimer);
        refreshTimer = null;
      }
      handshakeDone = false;
      historyLoaded = false;
      isConnecting = false;
      scheduleReconnect();
    };

    ws.onerror = () => {
      if (ws && ws.readyState !== WebSocket.CLOSED) {
        ws.close();
      }
    };
  }

  connectWebSocket();

  let activeUserIDs = [];
  const usersInfo = new Map();
  const conversations = new Map();
  const conversationOffsets = new Map();
  const olderMessagesLoadStateByUser = new Map();
  const notifiedUsers = new Set();
  let activeConversation = null;
  let handshakeDone = false;
  let refreshTimer = null;
  let historyLoaded = false;

  function nowString() {
    const d = new Date();
    return (
      d.getFullYear() +
      "-" +
      String(d.getMonth() + 1).padStart(2, "0") +
      "-" +
      String(d.getDate()).padStart(2, "0") +
      " " +
      String(d.getHours()).padStart(2, "0") +
      ":" +
      String(d.getMinutes()).padStart(2, "0") +
      ":" +
      String(d.getSeconds()).padStart(2, "0")
    );
  }

  function sendSystem(command) {
    if (!ws || ws.readyState !== WebSocket.OPEN) return;
    ws.send(`0|0|${nowString()}|${command}`);
  }

  function ensureUserInfo(id) {
    if (!usersInfo.has(id)) {
      sendSystem(`linkidtouser|${id}`);
    }
  }

  function renderUsersList() {
    listEl.innerHTML = "";

    if (!activeUserIDs.length && !currentUserID && conversations.size === 0) {
      const empty = document.createElement("div");
      empty.className = "user-item";
      empty.textContent = "Look like no one is online";
      listEl.appendChild(empty);
      return;
    }

    const usersWithMessages = Array.from(conversations.keys());

    usersWithMessages.sort((a, b) => {
      const msgsA = conversations.get(a) || [];
      const msgsB = conversations.get(b) || [];

      const lastA = msgsA.length
        ? new Date(msgsA[msgsA.length - 1].timestamp).getTime()
        : 0;
      const lastB = msgsB.length
        ? new Date(msgsB[msgsB.length - 1].timestamp).getTime()
        : 0;

      return lastB - lastA;
    });

    const otherActive = activeUserIDs.filter((id) => !conversations.has(id));

    otherActive.sort((a, b) => {
      const nameA = usersInfo.get(a)?.username || `User ${a}`;
      const nameB = usersInfo.get(b)?.username || `User ${b}`;
      return nameA.localeCompare(nameB);
    });

    const displayIDs = [...usersWithMessages, ...otherActive];

    if (currentUserID && !displayIDs.includes(currentUserID)) {
      displayIDs.unshift(currentUserID);
    }

    displayIDs.forEach((id) => {
      ensureUserInfo(id);

      const info = usersInfo.get(id);
      const el = document.createElement("div");
      el.className = "user-item";
      el.textContent = info ? info.username : `User ${id}`;

      if (id === currentUserID) {
        el.classList.add("is-self");
      }

      if (notifiedUsers.has(id)) {
        el.classList.add("has-notif");
      }

      el.dataset.userId = String(id);

      el.addEventListener("click", () => {
        if (id === currentUserID) return;
        openChatModal(id);
      });

      listEl.appendChild(el);
    });
  }

  function openChatModal(userID) {
    activeConversation = userID;
    if (notifiedUsers.has(userID)) {
      notifiedUsers.delete(userID);
      renderUsersList();
    }
    sendSystem(`markread|${userID}`);
    const info = usersInfo.get(userID);
    chatTitle.textContent = info ? info.username : `User ${userID}`;
    chatModal.classList.add("active");
    chatModal.setAttribute("aria-hidden", "false");
    renderConversation(userID);
    setTimeout(() => chatInput.focus(), 0);

    conversationOffsets.set(userID, 10);
  }

  function clearOlderMessagesLoadState(userID) {
    const loadState = olderMessagesLoadStateByUser.get(userID);
    if (!loadState) return;
    if (loadState.timer) {
      clearTimeout(loadState.timer);
    }
    olderMessagesLoadStateByUser.delete(userID);
  }

  function scheduleOlderMessagesLoadStateClear(userID) {
    const loadState = olderMessagesLoadStateByUser.get(userID);
    if (!loadState) return;
    if (loadState.timer) {
      clearTimeout(loadState.timer);
    }
    loadState.timer = setTimeout(() => {
      clearOlderMessagesLoadState(userID);
    }, 300);
  }

  // load older messages on scroll
  chatMessages.addEventListener("scroll", () => {
    setTimeout(() => {
    if (
      chatMessages.scrollTop <= 1 &&
      activeConversation &&
      !olderMessagesLoadStateByUser.has(activeConversation)
    ) {
      olderMessagesLoadStateByUser.set(activeConversation, {
        previousTop: chatMessages.scrollTop,
        previousHeight: chatMessages.scrollHeight,
        timer: null,
      });
      scheduleOlderMessagesLoadStateClear(activeConversation);

      const offset = conversationOffsets.get(activeConversation) || 10;
      sendSystem(`fetchmessages|${activeConversation}|${offset}`);
      conversationOffsets.set(activeConversation, offset + 10);
    }
    }, 500)
  });

  window.closeChatModal = function () {
    chatModal.classList.remove("active");
    chatModal.setAttribute("aria-hidden", "true");
    activeConversation = null;
  };

  chatModal.addEventListener("click", (event) => {
    if (event.target === chatModal) {
      window.closeChatModal();
    }
  });

  document.addEventListener("keydown", (event) => {
    if (event.key === "Escape" && chatModal.classList.contains("active")) {
      window.closeChatModal();
    }
  });

  function renderConversation(userID) {
    const loadState = olderMessagesLoadStateByUser.get(userID);
    const keepPositionAfterOlderLoad = Boolean(loadState);

    chatMessages.innerHTML = "";
    const msgs = (conversations.get(userID) || []).slice();
    msgs.sort((a, b) => {
      const ta = a.timestamp ? new Date(a.timestamp).getTime() : 0;
      const tb = b.timestamp ? new Date(b.timestamp).getTime() : 0;
      if (ta !== tb) return ta - tb;
      return (a._seq || 0) - (b._seq || 0);
    });
    msgs.forEach((msg) => {
      const bubble = document.createElement("div");
      const isOutgoing = msg.senderID === currentUserID;
      bubble.className = `chat-message ${isOutgoing ? "outgoing" : "incoming"}`;

      const header = document.createElement("div");
      header.className = "chat-message-header";

      const senderSpan = document.createElement("span");
      senderSpan.className = "chat-message-sender";
      if (isOutgoing) {
        senderSpan.textContent = "Me";
      } else {
        const info = usersInfo.get(msg.senderID);
        senderSpan.textContent = info ? info.username : `User ${msg.senderID}`;
      }

      const timeSpan = document.createElement("span");
      timeSpan.className = "chat-message-time";
      if (msg.timestamp) {
        const match = msg.timestamp.match(
          /^(\d{4}-\d{2}-\d{2})[T ](\d{2}:\d{2}:\d{2})/,
        );
        if (match) {
          const d = new Date(`${match[1]}T${match[2]}`);
          if (!isNaN(d.getTime())) {
            const dateStr = d.toLocaleDateString([], {
              day: "2-digit",
              month: "2-digit",
              year: "numeric",
            });
            const timeStr = d.toLocaleTimeString([], {
              hour: "2-digit",
              minute: "2-digit",
            });
            timeSpan.textContent = `${dateStr} ${timeStr}`;
          }
        }
      }

      header.appendChild(senderSpan);
      header.appendChild(timeSpan);

      // --- Contenu du message ---
      const contentDiv = document.createElement("div");
      contentDiv.className = "chat-message-content";
      contentDiv.textContent = msg.content;

      bubble.appendChild(header);
      bubble.appendChild(contentDiv);
      chatMessages.appendChild(bubble);
    });

    if (keepPositionAfterOlderLoad) {
      const heightDelta = chatMessages.scrollHeight - loadState.previousHeight;
      chatMessages.scrollTop = loadState.previousTop + Math.max(0, heightDelta);
    } else {
      chatMessages.scrollTop = chatMessages.scrollHeight;
    }
  }

  let messageSeq = 0;
  function addMessage(userID, message, silent = false) {
    const isContact = userID && userID !== currentUserID;
    const isIncoming = isContact && message.senderID !== currentUserID;
    if (isContact) {
      ensureUserInfo(userID);
    }
    if (isIncoming) {
      if (activeConversation === userID) {
        sendSystem(`markread|${userID}`);
      } else if (!silent) {
        notifiedUsers.add(userID);
      }
    }
    if (!conversations.has(userID)) {
      conversations.set(userID, []);
    }
    message._seq = messageSeq++;
    const msgs = conversations.get(userID);
    let inserted = false;
    for (let i = 0; i < msgs.length; i++) {
      if (
        new Date(message.timestamp).getTime() <
        new Date(msgs[i].timestamp).getTime()
      ) {
        msgs.splice(i, 0, message);
        inserted = true;
        break;
      }
    }
    if (!inserted) msgs.push(message);
    if (activeConversation === userID) {
      if (olderMessagesLoadStateByUser.has(userID)) {
        scheduleOlderMessagesLoadStateClear(userID);
      }
      renderConversation(userID);
    }
    renderUsersList();
  }

  function handleActiveUsers(payload) {
    const ids = payload
      .replace("activeusers:", "")
      .split(",")
      .map((id) => Number(id))
      .filter((id) => Number.isFinite(id) && id > 0);

    activeUserIDs = ids;

    ids.forEach((id) => {
      ensureUserInfo(id);
    });

    renderUsersList();
  }

  function handleUserInfo(payload) {
    const infoParts = payload.replace("userinfo:", "").split(",");
    const id = Number(infoParts[0]);
    if (!Number.isFinite(id)) return;

    usersInfo.set(id, {
      username: infoParts[1] || `User ${id}`,
      firstName: infoParts[2] || "",
      lastName: infoParts[3] || "",
      email: infoParts[4] || "",
    });

    renderUsersList();
  }

  function handleUnreadUsers(payload) {
    notifiedUsers.clear();

    payload
      .replace("unreadusers:", "")
      .split(",")
      .map((id) => Number(id))
      .filter((id) => Number.isFinite(id) && id > 0 && id !== currentUserID)
      .forEach((id) => notifiedUsers.add(id));

    renderUsersList();
  }

  chatForm.addEventListener("submit", (event) => {
    event.preventDefault();
    const content = chatInput.value.trim();
    if (!content || !activeConversation) return;
    if (!ws || ws.readyState !== WebSocket.OPEN) {
      connectWebSocket();
      return;
    }

    const timestamp = nowString();
    ws.send(`${currentUserID}|${activeConversation}|${timestamp}|${content}`);
    addMessage(activeConversation, {
      senderID: currentUserID,
      receiverID: activeConversation,
      timestamp,
      content,
    });
    chatInput.value = "";
  });
});
