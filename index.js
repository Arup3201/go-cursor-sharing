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
      const data = JSON.parse(message.data);
      if (!data.cursor) {
        console.error("cursor data missing");
        return;
      }

      console.log(data);

      const cursorData = data.cursor;
      switch (data.type) {
        case "remove":
          if (cursorData.id && cursorData.color) {
            const cursor = getOrCreateCursor(cursorData);
            cursor.remove();
          } else {
            console.error(
              "Cursor data needs id and color, but some are missing... " +
                cursorData,
            );
          }
          break;
        case "move":
          if (cursorData.id && cursorData.color) {
            const cursor = getOrCreateCursor(cursorData);
            cursor.style.transform = `translate(${cursorData.x}px, ${cursorData.y}px)`;
          } else {
            console.error(
              "Cursor data needs id and color, but some are missing... " +
                cursorData,
            );
          }
          break;
        default:
          console.log("Not supported message type");
      }
    } catch (err) {
      console.error("Error parsing message body, ", err);
    }
  };

  function getOrCreateCursor(cursorData) {
    const sender = cursorData.id;
    const existing = document.querySelector(`[data-sender='${sender}']`);
    if (existing) {
      return existing;
    }

    const template = document.getElementById("cursor");
    const cursor = template.content.firstElementChild.cloneNode(true);
    const svgPath = cursor.getElementsByTagName("path")[0];

    cursor.setAttribute("data-sender", sender);
    svgPath.setAttribute("fill", cursorData.color);
    document.body.appendChild(cursor);

    return cursor;
  }
})();
