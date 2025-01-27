const fs = require("fs");
const path = require("path");
const puppeteer = require("puppeteer");

const MAIN_LINK = "https://www.cms.ba.gov.br/transparencia/despesas-viagem";
const SCRIPT_TIME_LABEL = "Script Time";
const USER_AGENT =
  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36";
const PATH_FILES_FOLDER = "./travelExpensesFiles";

async function travelExpensesDataJob() {
  try {
    console.time(SCRIPT_TIME_LABEL);

    await checkAndCreateFolder(PATH_FILES_FOLDER);

    const [context, browser, page] = await initialConfigs();

    await page.goto(MAIN_LINK, { waitUntil: "networkidle0" });

    const pagesQuantity = await getTravelNumberPages(page);

    let travelExpensesList = [];

    for (let i = 1; i <= pagesQuantity; i++) {
      let url = `${MAIN_LINK}?page=${i}`;
      console.log(`Acessando URL: ${url}`);

      await page.goto(url, { waitUntil: "networkidle0", timeout: 0 });

      // Espera o seletor aparecer na página
      await page.waitForSelector(".audiencias_docs.list-none li", {
        timeout: 5000,
      });

      const pageDetails = await getAllTravelDetails(page);
      console.log(`Detalhes encontrados na página ${i}: ${pageDetails.length}`);

      travelExpensesList = travelExpensesList.concat(pageDetails);
    }

    console.log(travelExpensesList);

    await saveDataToJson(
      travelExpensesList,
      await getFormattedPath("./travelExpensesFiles/travelExpensesInfo.json")
    );

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
    headless: "new",
    defaultViewport: null,
    executablePath:
      process.env.PUPPETEER_EXECUTABLE_PATH || "/usr/bin/google-chrome",
  };

  const browser = await puppeteer.launch(options);

  const context = await browser.createIncognitoBrowserContext();
  const page = await context.newPage();

  await page.setUserAgent(USER_AGENT);

  await context.overridePermissions(MAIN_LINK, ["geolocation"]);

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

async function getAllTravelDetails(page) {
  const allDetails = await page.evaluate(() => {
    const detailsList = [];
    const elements = document.querySelectorAll(".audiencias_docs.list-none li");

    elements.forEach((element) => {
      const textContent = element.textContent.trim();

      // Usando regex para capturar os valores
      const info = {
        data: textContent.match(/Data:\s*(.*?)(?:\n|Tipo)/)?.[1]?.trim() || "",
        tipo:
          textContent.match(/Tipo:\s*(.*?)(?:\n|Usuário)/)?.[1]?.trim() || "",
        usuario:
          textContent.match(/Usuário:\s*(.*?)(?:\n|Valor)/)?.[1]?.trim() || "",
        valor:
          textContent
            .match(/Valor:\s*(.*?)(?:\n|Localidade)/)?.[1]
            ?.trim()
            .replace("R$ ", "") || "",
        localidade:
          textContent
            .match(/Localidade:\s*(.*?)(?:\n|Justificativa)/)?.[1]
            ?.trim() || "",
        justificativa:
          textContent.match(/Justificativa:\s*(.*)/)?.[1]?.trim() || "",
      };

      // Somente adiciona se a informação foi encontrada
      if (info.data) {
        detailsList.push(info);
      }
    });

    return detailsList;
  });

  return allDetails;
}

async function getTravelNumberPages(page) {
  const numberOfPages = await page.evaluate(() => {
    const paginationElement = document.querySelector(".pagination");
    const targetElement =
      paginationElement.children[paginationElement.children.length - 2];
    return parseInt(targetElement.textContent, 10);
  });

  console.log(`Total de páginas: ${numberOfPages}`);
  return numberOfPages;
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

async function getTimeNow() {
  const now = new Date();
  return await getFormattedDate(now);
}

async function saveDataToJson(data, filename) {
  fs.writeFileSync(filename, JSON.stringify(data, null, 2));
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

travelExpensesDataJob();
