const fs = require("fs");
const path = require("path");
const puppeteer = require("puppeteer");
const xml2js = require("xml2js");

// Update paths to use src directory
const DOWNLOAD_FOLDER_PATH = path.join(__dirname, "frequencyFiles");
const INPUT_PATH = path.join(DOWNLOAD_FOLDER_PATH, "LEG_SYS_frequencia.xml");
const OUTPUT_PATH = path.join(DOWNLOAD_FOLDER_PATH, "LEG_SYS_frequencia.json");
const DOWNLOAD_BUTTON_SELECTOR = ".scButton_default";
const EXPECTED_FILENAME = "LEG_SYS_frequencia.xml";
const LINK = "http://45.4.247.157/leg/salvador/LEG_SYS_frequencia/";
const USER_AGENT =
  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36";
const SCRIPT_TIME_LABEL = "Script Time";
const XML_BUTTON_SELECTOR = "#xml_top";
const PATH_FILES_FOLDER = path.join(__dirname, "frequencyFiles");

async function frequencyDataJob() {
  try {
    console.time(SCRIPT_TIME_LABEL);

    await checkAndCreateFolder(PATH_FILES_FOLDER);

    const [context, browser, page] = await initialConfigs();

    await goToXMLDownloadPage(page);

    await waitForAvailableDownload(page);

    await makeDownload(page);

    // Wait for download and get the actual downloaded file path
    const actualDownloadedFile = await waitForDownloadComplete(
      DOWNLOAD_FOLDER_PATH,
      EXPECTED_FILENAME
    ).catch((error) => {
      writeLog(error);
      throw error;
    });
    await writeLog(`Download concluído: ${actualDownloadedFile}`);

    // Dá um tempo extra para garantir que o arquivo esteja pronto
    await wait(3000);

    // Cria um novo caminho com timestamp para o arquivo XML
    const newXmlPath = await getFormattedPathForXml(actualDownloadedFile);

    // Renomeia o arquivo baixado e remove o original
    await renameDownloadedFile(actualDownloadedFile, newXmlPath);

    await writeLog(`Arquivo XML renomeado para: ${newXmlPath}`);

    // Use the renamed XML file path for conversion
    const jsonOutputPath = await getFormattedPath(OUTPUT_PATH);
    await convertXmlToJson(newXmlPath, jsonOutputPath)
      .then(() => writeLog(`Arquivo JSON convertido para: ${jsonOutputPath}`))
      .catch((error) => {
        writeLog(error);
        throw error;
      });

    await wait(5000);

    await browser.close();
    console.timeEnd(SCRIPT_TIME_LABEL);
  } catch (error) {
    await writeLog(error);
    process.exit();
  }
}

async function initialConfigs() {
  const myArgs = [
    "--disable-extensions",
    "--disable-features=IsolateOrigins,site-per-process",
    "--disable-gpu",
    "--disable-infobars",
    "--disable-setuid-sandbox",
    "--disable-web-security",
    "--enable-webgl",
    "--enable-accelerated-2d-canvas",
    "--force-device-scale-factor",
    "--ignore-certificate-errors",
    "--no-sandbox",
    "--disable-features=site-per-process",
    "--disable-features=IsolateOrigins,site-per-process,SitePerProcess",
    "--flag-switches-begin --disable-site-isolation-trials --flag-switches-end",
  ];

  const options = {
    args: myArgs,
    headless: false,
    // headless: "new",
    defaultViewport: null,
    executablePath:
      "C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe",
    // executablePath:
    //   process.env.PUPPETEER_EXECUTABLE_PATH || "/usr/bin/google-chrome",
  };

  const browser = await puppeteer.launch(options);

  const context = await browser.createIncognitoBrowserContext();
  const page = await context.newPage();

  await page.setUserAgent(USER_AGENT);

  const client = await page.target().createCDPSession();
  await client.send("Page.setDownloadBehavior", {
    behavior: "allow",
    downloadPath: DOWNLOAD_FOLDER_PATH,
  });

  //   await context.overridePermissions(LINK, ["geolocation"]);

  await page.setViewport({ width: 1280, height: 800 });

  page.setDefaultTimeout(61000);

  return [context, browser, page];
}

async function checkAndCreateFolder(folderPath) {
  const resolvedPath = path.resolve(folderPath);

  if (!fs.existsSync(resolvedPath)) {
    fs.mkdirSync(resolvedPath, { recursive: true });
    await writeLog(`Pasta '${resolvedPath}' criada com sucesso.`);
  } else {
    await writeLog(`Pasta '${resolvedPath}' já existe.`);
  }
}

async function convertXmlToJson(inputPath, outputPath) {
  const xml = fs.readFileSync(inputPath, "utf8");

  const parser = new xml2js.Parser({ explicitArray: false, mergeAttrs: true });

  const result = await parser.parseStringPromise(xml);

  const json = JSON.stringify(result, null, 2);

  fs.writeFileSync(outputPath, json, "utf8");

  await writeLog(`Arquivo JSON salvo em: ${outputPath}`);
}

async function getFormattedDate(date) {
  const options = {
    timeZone: "America/Sao_Paulo", // Configura o fuso horário para Brasília (BRT)
    hour12: false, // Usa formato de 24 horas
    // weekday: 'short', // Exibe apenas o dia da semana abreviado
    year: "numeric", // Exibe apenas o ano (com 4 dígitos)
    month: "2-digit", // Exibe o mês como dois dígitos
    day: "2-digit", // Exibe o dia do mês como dois dígitos
    hour: "2-digit", // Exibe a hora como dois dígitos
    minute: "2-digit", // Exibe os minutos como dois dígitos
    second: "2-digit", // Exibe os segundos como dois dígitos
  };

  return date.toLocaleString("pt-BR", options);
}

async function getFormattedPath(originalFilePath) {
  const now = new Date();

  const formattedDate = `${now.getFullYear()}${(now.getMonth() + 1)
    .toString()
    .padStart(2, "0")}${now.getDate().toString().padStart(2, "0")}`;
  const formattedTime = `${now.getHours().toString().padStart(2, "0")}${now
    .getMinutes()
    .toString()
    .padStart(2, "0")}${now.getSeconds().toString().padStart(2, "0")}`;

  const fileNameWithoutExtension = originalFilePath.replace(".json", "");

  const fileExtension = originalFilePath.split(".").pop();

  return `${fileNameWithoutExtension}_${formattedDate}_${formattedTime}.${fileExtension}`;
}

async function getFormattedPathForXml(originalFilePath) {
  const now = new Date();

  const formattedDate = `${now.getFullYear()}${(now.getMonth() + 1)
    .toString()
    .padStart(2, "0")}${now.getDate().toString().padStart(2, "0")}`;
  const formattedTime = `${now.getHours().toString().padStart(2, "0")}${now
    .getMinutes()
    .toString()
    .padStart(2, "0")}${now.getSeconds().toString().padStart(2, "0")}`;

  const directory = path.dirname(originalFilePath);
  const filename = path.basename(originalFilePath);

  const baseFilename = filename.split(".")[0];
  const fileExtension = path.extname(originalFilePath);

  return path.join(
    directory,
    `${baseFilename}_${formattedDate}_${formattedTime}${fileExtension}`
  );
}

async function renameDownloadedFile(oldPath, newPath) {
  return new Promise((resolve, reject) => {
    if (!fs.existsSync(oldPath)) {
      reject(new Error(`Arquivo não encontrado: ${oldPath}`));
      return;
    }

    // Copia o arquivo para o novo caminho com o timestamp
    fs.copyFile(oldPath, newPath, (err) => {
      if (err) {
        reject(err);
        return;
      }

      // Após cópia bem-sucedida, remove o arquivo original
      fs.unlink(oldPath, (unlinkErr) => {
        if (unlinkErr) {
          // Apenas loga o erro de remoção, mas não falha a operação
          console.log(
            `Aviso: Não foi possível remover o arquivo original: ${unlinkErr}`
          );
        }
        resolve();
      });
    });
  }).catch((error) => {
    throw new Error(`Erro ao processar o arquivo: ${error}`);
  });
}

async function goToXMLDownloadPage(page) {
  await page.goto(LINK, { waitUntil: "networkidle0" });

  await page.waitForSelector(XML_BUTTON_SELECTOR, { visible: true });
  await page.click(XML_BUTTON_SELECTOR);

  await wait(5000);
}

async function getTimeNow() {
  const now = new Date();
  return await getFormattedDate(now);
}

async function makeDownload(page) {
  let attempt = 0;
  while (attempt < 3) {
    try {
      await page.waitForSelector(DOWNLOAD_BUTTON_SELECTOR, {
        visible: true,
        timeout: 5000,
      });

      const downloadButton = await page.evaluateHandle(
        (DOWNLOAD_BUTTON_SELECTOR) =>
          Array.from(document.querySelectorAll(DOWNLOAD_BUTTON_SELECTOR)).find(
            (button) =>
              button.textContent.includes("Baixar") &&
              !button.classList.contains("disabled")
          ),
        DOWNLOAD_BUTTON_SELECTOR
      );

      if (downloadButton.asElement()) {
        await downloadButton.asElement().click();
        return;
      } else {
        throw new Error('Botão "Baixar" não encontrado.');
      }
    } catch (error) {
      attempt++;
      await writeLog(`Tentativa ${attempt} falhou. Tentando novamente...`);
      await wait(3000);
    }
  }
  throw new Error("Falha ao encontrar o botão de download após 3 tentativas.");
}

async function wait(time) {
  return new Promise(function (resolve) {
    setTimeout(resolve, time);
  });
}

async function waitForAvailableDownload(page) {
  await page.waitForFunction(
    () => {
      const exportMessages = Array.from(
        document.querySelectorAll(".scExportLineFont")
      );
      return exportMessages.some((message) =>
        message.textContent.includes("Arquivo gerado com sucesso")
      );
    },
    { timeout: 0 }
  );
}

async function waitForDownloadComplete(
  downloadPath,
  expectedFilename,
  timeout = 120000
) {
  const startTime = new Date().getTime();
  let foundFile = false;
  let filePath = null;

  await writeLog("Aguardando conclusão do download...");
  await writeLog(`Diretório de download: ${downloadPath}`);

  while (!foundFile) {
    // Check if timeout has been reached
    if (new Date().getTime() - startTime > timeout) {
      throw new Error(`Download timeout após ${timeout / 1000} segundos`);
    }

    try {
      const files = fs.readdirSync(downloadPath);

      // Check for temp download files first (Chrome creates .crdownload files)
      const downloadingFiles = files.filter((file) =>
        file.endsWith(".crdownload")
      );
      if (downloadingFiles.length > 0) {
        await writeLog("Download ainda em progresso...");
        await wait(2000);
        continue;
      }

      // Now look for the actual XML file
      for (const file of files) {
        if (file.includes("LEG_SYS_frequencia") && file.endsWith(".xml")) {
          filePath = path.join(downloadPath, file);

          // Ensure file is fully written and not empty
          const stats = fs.statSync(filePath);
          if (stats.size > 0) {
            // Double check file is not changing in size
            await wait(2000);
            const newStats = fs.statSync(filePath);
            if (stats.size === newStats.size) {
              foundFile = true;
              await writeLog(
                `Arquivo encontrado: ${file} (${stats.size} bytes)`
              );
              break;
            }
          }
        }
      }

      if (!foundFile) {
        await writeLog("Arquivo ainda não encontrado, aguardando...");
        await wait(3000);
      }
    } catch (err) {
      await writeLog(`Erro ao verificar arquivos: ${err.message}`);
      await wait(2000);
    }
  }

  return filePath;
}

async function writeLog(receivedString) {
  let string = `${await getTimeNow()} - `;
  if (typeof receivedString === "object" && receivedString !== null) {
    if (receivedString instanceof Error) {
      string += `Error: ${receivedString.message}\nStack: ${receivedString.stack}`;
    } else {
      string += `Object: ${JSON.stringify(receivedString)}`;
    }
  } else {
    string += receivedString;
  }
  string += "\n";

  console.log(string);
}

frequencyDataJob();
