// Minimal front-end chat client with streaming + thinking placeholder

const $ = (sel) => document.querySelector(sel);
const chatLog = $('#chat-log');
const form = $('#composer-form');
const input = $('#user-input');
const sendBtn = $('#send-btn');
const reloadBtn = $('#reload-btn');
const statusEl = $('#status-indicator');
const progress = $('#progress');

const API = {
  chat: '/chat/',
  chatStream: '/chat/stream/',
  reload: '/reload-model/',
  health: '/health/',
};

function el(tag, className, children = []) {
  const node = document.createElement(tag);
  if (className) node.className = className;
  for (const child of children) node.append(child);
  return node;
}

function scrollToBottom() {
  chatLog.scrollTop = chatLog.scrollHeight;
}

function addUserMessage(text) {
  const msg = el('div', 'msg user', [
    el('div', 'avatar', ['🧑']),
    el('div', 'bubble', [text])
  ]);
  chatLog.append(msg);
  scrollToBottom();
}

function addAssistantMessage(text) {
  const msg = el('div', 'msg bot', [
    el('div', 'avatar', ['🤖']),
    el('div', 'bubble', [text])
  ]);
  chatLog.append(msg);
  scrollToBottom();
}

function addThinkingPlaceholder() {
  const thinking = el('div', 'msg bot');
  const avatar = el('div', 'avatar', ['🤖']);
  const bubble = el('div', 'bubble');
  const inline = el('span', 'thinking', ['Thinking', el('span', 'dots')]);
  bubble.append(inline);
  thinking.append(avatar, bubble);
  chatLog.append(thinking);
  scrollToBottom();
  return { root: thinking, bubble, inline };
}

function setStatus(text) {
  statusEl.textContent = text || '';
}

function showProgress(show) {
  progress.classList.toggle('hidden', !show);
}

async function reloadModel() {
  showProgress(true);
  setStatus('Reloading model...');
  try {
    const res = await fetch(API.reload, { method: 'POST' });
    const data = await res.json().catch(() => ({}));
    if (!res.ok) throw new Error(data.error || 'Failed to reload');
    setStatus('Model reloaded');
  } catch (err) {
    setStatus('Reload failed');
    addAssistantMessage(`⚠️ ${err.message}`);
  } finally {
    showProgress(false);
    setTimeout(() => setStatus(''), 1500);
  }
}

async function sendMessage(text) {
  if (!text || !text.trim()) return;
  const prompt = text.trim();

  addUserMessage(prompt);
  input.value = '';
  input.focus();

  // Insert thinking placeholder and progress
  const placeholder = addThinkingPlaceholder();
  showProgress(true);
  setStatus('Thinking...');

  // Stream from backend (NDJSON lines from Ollama)
  let acc = '';
  try {
    const res = await fetch(API.chatStream, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ prompt })
    });

    if (!res.ok || !res.body) {
      // Fallback to non-streaming endpoint
      const fallback = await fetch(API.chat, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ prompt })
      });
      const data = await fallback.json();
      if (!fallback.ok) throw new Error(data.error || 'Request failed');
      acc = data.response || '';
      placeholder.bubble.textContent = acc || '(no response)';
      return;
    }

    const reader = res.body.getReader();
    const decoder = new TextDecoder();
    let buf = '';

    while (true) {
      const { value, done } = await reader.read();
      if (done) break;
      buf += decoder.decode(value, { stream: true });

      let idx;
      while ((idx = buf.indexOf('\n')) >= 0) {
        const line = buf.slice(0, idx).trim();
        buf = buf.slice(idx + 1);
        if (!line) continue;
        try {
          const obj = JSON.parse(line);
          if (obj.response) {
            acc += obj.response;
            placeholder.bubble.textContent = acc;
          }
          if (obj.done) {
            // stream finished
          }
        } catch (_) {
          // ignore malformed line
        }
      }
    }
  } catch (err) {
    placeholder.bubble.classList.add('error');
    placeholder.bubble.textContent = `Error: ${err.message}`;
  } finally {
    // Hide progress, clear status
    showProgress(false);
    setStatus('');
  }
}

async function init() {
  // Health ping (soft)
  try {
    const res = await fetch(API.health);
    if (res.ok) setStatus('Ready');
  } catch (_) {}
  setTimeout(() => setStatus(''), 1200);

  form.addEventListener('submit', (e) => {
    e.preventDefault();
    sendMessage(input.value);
  });

  input.addEventListener('keydown', (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      form.dispatchEvent(new Event('submit'));
    }
  });

  reloadBtn?.addEventListener('click', reloadModel);
}

window.ChatUI = { addAssistantMessage };

window.addEventListener('DOMContentLoaded', init);

