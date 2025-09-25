export async function initWebRTC(
  roomId: string,
  wsUrl: string,
  onTrack: (stream: MediaStream) => void
) {
  const pc = new RTCPeerConnection({
    iceServers: [{ urls: "stun:stun.l.google.com:19302" }],
  });

  const ws = new WebSocket(wsUrl);

  // Posíláme ICE kandidáty na server
  pc.onicecandidate = (event) => {
    if (event.candidate) {
      ws.send(
        JSON.stringify({ type: "candidate", candidate: event.candidate, room: roomId })
      );
    }
  };

  // Když přijde vzdálený audio stream
  pc.ontrack = (event) => {
    onTrack(event.streams[0]);
  };

  // Lokální mikrofon
  const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
  stream.getTracks().forEach((track) => pc.addTrack(track, stream));

  ws.onmessage = async (event) => {
    const msg = JSON.parse(event.data);

    if (msg.type === "offer") {
      await pc.setRemoteDescription(new RTCSessionDescription(msg.offer));
      const answer = await pc.createAnswer();
      await pc.setLocalDescription(answer);
      ws.send(JSON.stringify({ type: "answer", answer, room: roomId }));
    } else if (msg.type === "answer") {
      await pc.setRemoteDescription(new RTCSessionDescription(msg.answer));
    } else if (msg.type === "candidate") {
      try {
        await pc.addIceCandidate(msg.candidate);
      } catch (err) {
        console.error("Error adding ICE candidate", err);
      }
    }
  };

  ws.onopen = async () => {
    const offer = await pc.createOffer();
    await pc.setLocalDescription(offer);
    ws.send(JSON.stringify({ type: "offer", offer, room: roomId }));
  };

  return { pc, ws };
}
