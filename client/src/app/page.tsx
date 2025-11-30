"use client";

import { useEffect } from "react";
import { useSignalingStore } from "@/store/useSignalingStore";

export default function Home() {
  const { connect, isConnected, sendMessage, messages } = useSignalingStore();

  useEffect(() => {
    connect();
    // Cleanup al desmontar componente no es estrictamente necesario en strict mode
    // para este singleton, pero es buena práctica si la conexión fuera por página.
  }, [connect]);

  return (
    <main className="flex min-h-screen flex-col items-center justify-center p-24 bg-gray-950 text-white">
      <h1 className="text-4xl font-bold mb-8">P2P Transfer Core</h1>

      <div className="flex gap-4 mb-8">
        <div
          className={`w-4 h-4 rounded-full ${
            isConnected ? "bg-green-500" : "bg-red-500"
          }`}
        />
        <span>
          Estado: {isConnected ? "Conectado al Servidor" : "Desconectado"}
        </span>
      </div>

      <button
        onClick={() =>
          sendMessage({ type: "ping", payload: "Hola desde el cliente!" })
        }
        className="px-6 py-3 bg-blue-600 rounded-lg hover:bg-blue-700 transition font-medium"
      >
        Enviar Ping
      </button>

      <div className="mt-8 w-full max-w-md bg-gray-900 p-4 rounded-lg border border-gray-800">
        <h3 className="text-lg font-semibold mb-2 text-gray-400">
          Log de Mensajes:
        </h3>
        <ul className="space-y-2 font-mono text-sm h-64 overflow-y-auto">
          {messages.map((msg, i) => (
            <li key={i} className="break-all border-b border-gray-800 pb-1">
              {msg}
            </li>
          ))}
        </ul>
      </div>
    </main>
  );
}
