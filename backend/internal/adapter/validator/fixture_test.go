package validator

import "strings"

const configBlock = `    const CONFIG = {
      tempo: 5,
      palette: ["#101018", "#3a3a55", "#39ff14"],
      mapSize: "medium",
      difficulty: "normal",
      controls: "touch+keyboard",
      style: "neon"
    };
`

const validGameHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>Neon Snake</title>
  <style>
    body { margin: 0; background: #101018; color: #e8e8f0; font-family: system-ui, sans-serif; }
    #wrap { display: flex; flex-direction: column; align-items: center; gap: 10px; padding: 20px; }
    #board { border: 2px solid #3a3a55; border-radius: 6px; }
  </style>
</head>
<body>
  <div id="wrap">
    <h1>Neon Snake</h1>
    <div id="score">Score: 0</div>
    <canvas id="board" width="320" height="320"></canvas>
  </div>
  <script>
` + configBlock + `
    const canvas = document.getElementById("board");
    const ctx = canvas.getContext("2d");
    const cell = 20;
    const cols = 16;
    let snake = [{ x: 4, y: 4 }];
    let dir = { x: 1, y: 0 };
    let food = { x: 10, y: 10 };
    let score = 0;

    function draw() {
      ctx.fillStyle = CONFIG.palette[0];
      ctx.fillRect(0, 0, canvas.width, canvas.height);
      ctx.fillStyle = CONFIG.palette[2];
      for (const part of snake) {
        ctx.fillRect(part.x * cell, part.y * cell, cell - 2, cell - 2);
      }
      ctx.fillStyle = CONFIG.palette[1];
      ctx.fillRect(food.x * cell, food.y * cell, cell - 2, cell - 2);
      document.getElementById("score").textContent = "Score: " + score;
    }

    function step() {
      const head = snake[0];
      const next = { x: (head.x + dir.x + cols) % cols, y: (head.y + dir.y + cols) % cols };
      if (next.x === food.x && next.y === food.y) {
        score = score + 1;
        food = { x: Math.floor(Math.random() * cols), y: Math.floor(Math.random() * cols) };
      } else {
        snake.pop();
      }
      snake.unshift(next);
      draw();
    }

    canvas.addEventListener("touchstart", function (event) {
      event.preventDefault();
      const touch = event.changedTouches[0];
      dir = touch.clientX > canvas.width / 2 ? { x: 1, y: 0 } : { x: -1, y: 0 };
    });

    document.addEventListener("keydown", function (event) {
      if (event.key === "ArrowUp") { dir = { x: 0, y: -1 }; }
      if (event.key === "ArrowDown") { dir = { x: 0, y: 1 }; }
    });

    draw();
    setInterval(step, 1000 / CONFIG.tempo / 2);
  </script>
</body>
</html>
`

func gameWithExternalURL() string {
	return strings.Replace(validGameHTML, "  <script>",
		"  <script src=\"https://cdn.example.com/x.js\"></script>\n  <script>", 1)
}

func gameWithoutConfig() string {
	return strings.Replace(validGameHTML, configBlock, "", 1)
}

func gameWithForbiddenAPI() string {
	return strings.Replace(validGameHTML, "    let score = 0;",
		"    let score = Number(localStorage.getItem(\"score\") || 0);", 1)
}

func gameWithBrokenJS() string {
	return strings.Replace(validGameHTML, "  </script>", "    function(){\n  </script>", 1)
}
