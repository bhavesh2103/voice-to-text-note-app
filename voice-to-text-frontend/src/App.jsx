import { useEffect, useRef, useState } from "react";

function App() {
  const [isRecording, setIsRecording] = useState(false);
  const [recordings, setRecordings] = useState([]);
  const [devices, setDevices] = useState([]);
  const [selectedDeviceId, setSelectedDeviceId] = useState("");

  const mediaRecorderRef = useRef(null);
  const audioChunksRef = useRef([]);

  // 🎙 Load microphones on mount
  useEffect(() => {
    navigator.mediaDevices.getUserMedia({ audio: true }).then(() => {
      navigator.mediaDevices.enumerateDevices().then((deviceInfos) => {
        const mics = deviceInfos.filter((d) => d.kind === "audioinput");
        setDevices(mics);
        if (mics.length > 0) {
          setSelectedDeviceId(mics[0].deviceId);
        }
      });
    }).catch((err) => {
      console.error("Mic access error:", err);
    });
  }, []);

  // 🔴 Start recording
  const startRecording = async () => {
    const stream = await navigator.mediaDevices.getUserMedia({
      audio: { deviceId: selectedDeviceId },
    });

    mediaRecorderRef.current = new MediaRecorder(stream);
    audioChunksRef.current = [];

    mediaRecorderRef.current.ondataavailable = (e) => {
      audioChunksRef.current.push(e.data);
    };

    mediaRecorderRef.current.onstop = () => {
      const audioBlob = new Blob(audioChunksRef.current, { type: "audio/webm" });
      const audioUrl = URL.createObjectURL(audioBlob);
      const newRecording = { blob: audioBlob, url: audioUrl, transcript: null };

      setRecordings((prev) => {
        const updated = [...prev, newRecording];
        return updated.length > 3 ? updated.slice(updated.length - 3) : updated;
      });
    };

    mediaRecorderRef.current.start();
    setIsRecording(true);
  };

  // 🛑 Stop recording
  const stopRecording = () => {
    mediaRecorderRef.current.stop();
    setIsRecording(false);
  };

  // 📤 Transcribe selected recording
  const transcribeRecording = async (index) => {
    const rec = recordings[index];
    if (!rec || !rec.blob) return;

    const formData = new FormData();
    formData.append("audio", rec.blob, "note.webm");

    updateRecording(index, { transcript: "Transcribing..." });

    try {
      const res = await fetch("http://localhost:8080/transcribe", {
        method: "POST",
        body: formData,
      });

      const data = await res.json();
      updateRecording(index, {
        transcript: data.transcript || "No transcript found.",
      });
    } catch (err) {
      updateRecording(index, { transcript: "Error: " + err.message });
    }
  };

  // Update recording info
  const updateRecording = (index, updatedFields) => {
    setRecordings((prev) =>
      prev.map((r, i) => (i === index ? { ...r, ...updatedFields } : r))
    );
  };

  return (
    <div className="min-h-screen bg-gray-100 flex flex-col items-center p-6">
      <h1 className="text-3xl font-bold mb-6">🎤 Voice-to-Text Note App</h1>

      {/* 🎚 Mic Selector */}
      <div className="mb-4 w-full max-w-xl">
        <label className="font-semibold mr-2">Select Microphone:</label>
        <select
          className="px-3 py-1 border rounded w-full"
          value={selectedDeviceId}
          onChange={(e) => setSelectedDeviceId(e.target.value)}
        >
          {devices.length === 0 ? (
            <option disabled>No microphones found</option>
          ) : (
            devices.map((device) => (
              <option key={device.deviceId} value={device.deviceId}>
                {device.label || "Unnamed Microphone"}
              </option>
            ))
          )}
        </select>
      </div>

      {/* 🎙 Record Button */}
      <button
        onClick={isRecording ? stopRecording : startRecording}
        className={`px-6 py-2 rounded-lg text-white text-lg mb-6 ${
          isRecording ? "bg-red-500" : "bg-blue-600"
        }`}
      >
        {isRecording ? "Stop & Save" : "Start Recording"}
      </button>

      {/*  Recordings Preview */}
      <div className="w-full max-w-xl space-y-4">
        {recordings.length === 0 && (
          <div className="text-gray-500 text-center">
            No recordings yet. Click "Start Recording" to begin.
          </div>
        )}

        {recordings.map((rec, i) => (
          <div key={i} className="bg-white p-4 rounded shadow flex flex-col gap-2">
            <audio controls src={rec.url} className="w-full" />

            {rec.transcript ? (
              <div>
                <span className="font-semibold">Transcript:</span>{" "}
                <span className="text-gray-800">{rec.transcript}</span>
              </div>
            ) : (
              <button
                onClick={() => transcribeRecording(i)}
                className="self-start px-4 py-1 bg-green-600 text-white rounded"
              >
                Transcribe Now
              </button>
            )}
          </div>
        ))}
      </div>
    </div>
  );
}

export default App;
