const evtSource = new EventSource("/sse");

const box = document.getElementById("notif-box");
const text = document.getElementById("notif-text");
const closeBtn = document.getElementById("notif-close");

let timer = null;

function show(msg) {
  if (!msg) return;

  const formatted = formatNotif(msg);
  text.textContent = formatted;
  box.classList.add("active");

  if (timer) clearTimeout(timer);

  timer = setTimeout(() => hide(), 5000);
}

function hide() {
  box.classList.remove("active");
  if (timer) clearTimeout(timer);
  timer = null;
}

closeBtn.onclick = hide;

box.onmouseenter = () => {
  if (timer) clearTimeout(timer);
};

box.onmouseleave = () => {
  timer = setTimeout(hide, 5000);
};

function formatNotif(str) {
  const parts = str.split("|", 6);
  if (parts.length < 6) return str;

  const [sender, date, type, subjectType, subjectID, label] = parts;
  const dt = new Date(date);
  const hours = dt.getHours().toString().padStart(2, "0");
  const minutes = dt.getMinutes().toString().padStart(2, "0");

  let action = "";
  switch (type) {
    case "like":
      action = "has liked";
      break;
    case "dislike":
      action = "has disliked";
      break;
    case "comment":
      action = "has commented on";
      break;
    case "message":
      action = "sent you a message";
      break;
    default:
      action = type;
  }

  let maxLen = 25;
  let displayLabel =
    label.length > maxLen ? label.slice(0, maxLen) + "…" : label;

  let targetText =
    subjectType === "post"
      ? `your post "${displayLabel}"`
      : subjectType === "comment"
        ? `your comment (#${subjectID}): "${displayLabel}"`
        : subjectType;

  return `${sender} ${action} ${targetText} today at ${hours}:${minutes}`;
}

evtSource.onmessage = (e) => {
  show(e.data);
};
