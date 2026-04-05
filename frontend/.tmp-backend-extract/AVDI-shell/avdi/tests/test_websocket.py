#!/usr/bin/env python3
import asyncio
import websockets
import sys


async def test():
    uri = "ws://localhost:8081/ws"
    try:
        async with websockets.connect(uri) as websocket:
            print("Connected to WebSocket server")
            # Send a subscription message
            subscribe_msg = '{"action": "subscribe", "channel": "broadcast"}'
            await websocket.send(subscribe_msg)
            print(f"Sent: {subscribe_msg}")
            # Wait for a response (timeout 2 seconds)
            try:
                response = await asyncio.wait_for(websocket.recv(), timeout=2)
                print(f"Received: {response}")
            except asyncio.TimeoutError:
                print("No message received within 2 seconds")
            # Close
            await websocket.close()
            print("Connection closed")
    except Exception as e:
        print(f"Error: {e}")
        sys.exit(1)


if __name__ == "__main__":
    asyncio.run(test())
