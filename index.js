(async function () {
  function connect() {
    const ws = new WebSocket("ws://localhost:8081/ws");
    return new Promise((resolve, reject) => {
      const timer = setTimeout(() => {
        if (ws.readyState === 1) {
          clearInterval(timer);
          resolve(ws);
        }
      }, 10);
    });
  }

  const connection = await connect();

  document.body.onmousemove = (evt) => {
    const messageBody = { x: evt.clientX, y: evt.clientY };
    connection.send(JSON.stringify(messageBody));
  };

  connection.onmessage = (message) => {
    try {
      console.log(message.data);
      const messageBody = JSON.parse(message.data);
      if (messageBody.id && messageBody.color) {
        const cursor = getOrCreateCursor(messageBody);
        cursor.style.transform = `translate(${messageBody.x}px, ${messageBody.y}px)`;
      } else {
        console.error(
          "Message body is missing id or color of cursor, " + messageBody,
        );
      }
    } catch (err) {
      console.error("Error parsing message body, ", err);
    }
  };

  function getOrCreateCursor(messageBody) {
    const sender = messageBody.id;
    const existing = document.querySelector(`[data-sender='${sender}']`);
    if (existing) {
      return existing;
    }

    const template = document.getElementById("cursor");
    const cursor = template.content.firstElementChild.cloneNode(true);
    const svgPath = cursor.getElementsByTagName("path")[0];

    cursor.setAttribute("data-sender", sender);
    console.log(messageBody.color);
    svgPath.setAttribute("fill", messageBody.color);
    document.body.appendChild(cursor);

    return cursor;
  }
})();
