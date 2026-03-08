import json
import asyncio
from typing import Optional
from contextlib import AsyncExitStack

from mcp import ClientSession, StdioServerParameters
from mcp.client.stdio import stdio_client
from openai import AsyncOpenAI

# USING OPENAI COMPATIBLE API
API_KEY = ""
BASE_URL = "https://ai.sumopod.com/v1"
MODEL = "gpt-5-nano"

class MCPClient:
    def __init__(self):
        self.session: Optional[ClientSession] = None
        self.exit_stack = AsyncExitStack()
        self.llm_tools = []
        self.llm = None

    async def connect(self):
        server_params = StdioServerParameters(
            command="telbot", args=["--mcp"], env=None
        )
        stdio_transport = await self.exit_stack.enter_async_context(
            stdio_client(server_params)
        )
        self.stdio, self.write = stdio_transport
        self.session = await self.exit_stack.enter_async_context(
            ClientSession(self.stdio, self.write)
        )
        await self.session.initialize()

        response = await self.session.list_tools()
        tools = response.tools
        print("MCP tools:", [t.name for t in tools])

        for tool in tools:
            schema = tool.inputSchema if isinstance(tool.inputSchema, dict) else {}
            params = {"type": "object", "properties": schema.get("properties", {})}
            if "required" in schema:
                params["required"] = schema["required"]
            self.llm_tools.append(
                {
                    "type": "function",
                    "function": {
                        "name": tool.name,
                        "description": tool.description,
                        "parameters": params,
                    },
                }
            )

        self.llm = AsyncOpenAI(base_url=BASE_URL, api_key=API_KEY)

    async def chat(self, prompt: str):
        messages = [{"role": "user", "content": prompt}]

        response = await self.llm.chat.completions.create(
            model=MODEL, messages=messages, tools=self.llm_tools
        )

        message = response.choices[0].message
        if not message.tool_calls:
            print(message.content)
            return

        messages.append(message)
        for tc in message.tool_calls:
            args = json.loads(tc.function.arguments)
            print(f"[tool] {tc.function.name}({args})")
            result = await self.session.call_tool(tc.function.name, arguments=args)
            messages.append(
                {
                    "role": "tool",
                    "tool_call_id": tc.id,
                    "content": result.content[0].text,
                }
            )

        final = await self.llm.chat.completions.create(model=MODEL, messages=messages)
        print(f"\n{final.choices[0].message.content}")

    async def cleanup(self):
        await self.exit_stack.aclose()


async def main():
    client = MCPClient()
    try:
        await client.connect()
        print("\nReady! Type your prompt (or 'exit' to quit)\n")
        while True:
            prompt = input("> ")
            if prompt.strip().lower() in ("exit", "quit", "q"):
                break
            if not prompt.strip():
                continue
            await client.chat(prompt)
            print()
    finally:
        await client.cleanup()


if __name__ == "__main__":
    asyncio.run(main())
