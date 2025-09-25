import { useRef, useState } from "react";
import { initWebRTC } from "./webrtc";

const WS_URL = "ws://188.246.96.246:5173/ws";

export default function App() {
  const [roomId, setRoomId] = useState("");
  const [joined, setJoined] = useState(false);
  const audioRef = useRef<HTMLAudioElement>(null);

  async function joinRoom() {
    await initWebRTC(roomId, WS_URL, (remoteStream) => {
      if (audioRef.current) {
        audioRef.current.srcObject = remoteStream;
        audioRef.current.play().catch(console.error);
      }
    });
    setJoined(true);
  }

  return (
    <div className="card">
      {!joined ? (
        <>
          <h1>Join a Room</h1>
          <input
            type="text"
            placeholder="Room ID"
            value={roomId}
            onChange={(e) => setRoomId(e.target.value)}
          />
          <button onClick={joinRoom}>Join</button>
        </>
      ) : (
        <>
          <h1>Connected</h1>
          <p>Room: <strong>{roomId}</strong></p>
          <audio ref={audioRef} autoPlay className="hidden" />
          <div className="status">🎤 Audio is live</div>
        </>
      )}
    </div>
  );
}
