// Testes ponta a ponta. O Playwright sobe o próprio servidor (go run) num
// banco temporário, em modo development para ler templates e CSS do disco.
import { defineConfig, devices } from "@playwright/test";
import os from "node:os";
import path from "node:path";

const port = 8090;
const db = path.join(os.tmpdir(), `weeklly-e2e-${process.pid}.db`);
// Compila e executa o binário direto: com `go run` o servidor é neto do shell
// e sobrevive ao fim dos testes no Windows.
const bin = path.join("e2e", ".bin", process.platform === "win32" ? "weeklly.exe" : "weeklly");

export default defineConfig({
  testDir: "./tests",
  outputDir: "./test-results",
  fullyParallel: false,
  workers: 1,
  reporter: [["list"]],
  timeout: 30_000,
  use: {
    baseURL: `http://127.0.0.1:${port}`,
    locale: "pt-BR",
    timezoneId: "America/Sao_Paulo",
    colorScheme: "dark",
    trace: "retain-on-failure",
  },
  projects: [
    { name: "desktop", use: { ...devices["Desktop Chrome"], viewport: { width: 1440, height: 900 } } },
    { name: "mobile", use: { ...devices["Pixel 7"] } },
  ],
  webServer: {
    command: `go build -o ${bin} ./cmd/weeklly && ${bin}`,
    cwd: "..",
    env: {
      WEEKLLY_ENV: "development",
      WEEKLLY_ADDR: `127.0.0.1:${port}`,
      WEEKLLY_DB_PATH: db,
    },
    url: `http://127.0.0.1:${port}/healthz`,
    reuseExistingServer: false,
    timeout: 90_000,
  },
});
