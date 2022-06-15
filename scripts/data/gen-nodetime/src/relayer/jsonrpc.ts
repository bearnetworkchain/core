import { stdin } from "process";
import { JSONRPCServer, SimpleJSONRPCMethod} from "json-rpc-2.0";

type Handler = [string, SimpleJSONRPCMethod];

// run 通過標準流公開給定處理程序的 JSON-RPC 服務器。
export default async function run(handlers: Handler[]) {
  // 初始化 RPC 服務器。
  const server = new JSONRPCServer();

  // 將方法附加到 rpc 服務器。
  for (const [name, func] of handlers) {
    server.addMethod(name, func);
  }

  // read the rpc call, invoke it and send a response.
  let jsonRequest: string = "";

  stdin.setEncoding("utf8");

  for await (const chunk of stdin) {
    jsonRequest += chunk;
  }

  const jsonResponse = await server.receiveJSON(jsonRequest);
  const response = JSON.stringify(jsonResponse);

  console.log(response);
}

