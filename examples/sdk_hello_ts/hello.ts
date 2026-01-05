// TypeScript client for Gas Town REST/WebSocket API
// Run the server first: gt serve --runtime sdk
// Then run this: npx tsx examples/sdk_hello_ts/hello.ts

import WebSocket from "ws";

const API_BASE = process.env.GASTOWN_API || "http://localhost:8080";
const WS_BASE = API_BASE.replace(/^http/, "ws");

interface SessionResponse {
  session_id: string;
  agent_id: string;
  role: string;
  rig_name?: string;
  worker_name?: string;
  running: boolean;
  started_at: string;
  runtime_type: string;
}

interface WSMessage {
  type: "text" | "tool_call" | "tool_result" | "error" | "complete";
  content?: string;
  timestamp: string;
  error?: string;
}

async function createSession(): Promise<SessionResponse> {
  const response = await fetch(`${API_BASE}/sessions`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      agent_id: "example/hello-ts",
      role: "polecat",
      rig_name: "example",
      worker_name: "hello-ts",
      system_prompt:
        "You are a helpful programming assistant. Respond concisely.",
    }),
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(`Failed to create session: ${error.error}`);
  }

  return response.json();
}

async function sendPrompt(sessionId: string, prompt: string): Promise<void> {
  const response = await fetch(`${API_BASE}/sessions/${sessionId}/prompt`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ prompt }),
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(`Failed to send prompt: ${error.error}`);
  }
}

async function deleteSession(sessionId: string): Promise<void> {
  const response = await fetch(`${API_BASE}/sessions/${sessionId}`, {
    method: "DELETE",
  });

  if (!response.ok && response.status !== 204) {
    const error = await response.json();
    throw new Error(`Failed to delete session: ${error.error}`);
  }
}

function connectWebSocket(sessionId: string): Promise<WebSocket> {
  return new Promise((resolve, reject) => {
    const ws = new WebSocket(`${WS_BASE}/sessions/${sessionId}/ws`);

    ws.on("open", () => resolve(ws));
    ws.on("error", (err) => reject(new Error(`WebSocket connection failed: ${err.message}`)));

    // Set a connection timeout
    setTimeout(() => {
      if (ws.readyState !== WebSocket.OPEN) {
        ws.close();
        reject(new Error("WebSocket connection timeout"));
      }
    }, 5000);
  });
}

async function main() {
  console.log("Gas Town API Client Example");
  console.log("===========================\n");
  console.log(`Connecting to: ${API_BASE}\n`);

  // Create a session
  console.log("Creating session...");
  const session = await createSession();
  console.log(`Session created: ${session.session_id}`);
  console.log(`Runtime: ${session.runtime_type}\n`);

  // Connect WebSocket for streaming responses
  console.log("Connecting WebSocket...");
  const ws = await connectWebSocket(session.session_id);
  console.log("WebSocket connected\n");

  // Set up message handler
  const responsePromise = new Promise<void>((resolve, reject) => {
    ws.on("message", (data) => {
      const msg: WSMessage = JSON.parse(data.toString());

      switch (msg.type) {
        case "text":
          process.stdout.write(msg.content || "");
          break;
        case "error":
          console.error(`\nError: ${msg.error}`);
          resolve();
          break;
        case "complete":
          console.log("\n");
          resolve();
          break;
      }
    });

    ws.on("error", (err) => reject(new Error(`WebSocket error: ${err.message}`)));
    ws.on("close", () => resolve());
  });

  // Send prompt
  const prompt = "Write a Hello World program in Ada. Just the code, no explanation.";
  console.log(`Prompt: ${prompt}\n`);
  console.log("Response:");
  console.log("─────────");

  await sendPrompt(session.session_id, prompt);

  // Wait for response to complete
  await responsePromise;

  console.log("─────────");
  console.log("Done!");

  // Cleanup
  ws.close();
  await deleteSession(session.session_id);
  console.log("Session deleted");
}

main().catch((err) => {
  console.error("Error:", err.message);
  process.exit(1);
});
