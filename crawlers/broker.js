const amqp = require("amqplib");

async function sendMessage(queue, message) {
  try {
    console.log("Tentando conectar ao RabbitMQ em:", process.env.BROKER_URL);
    const connection = await amqp.connect(process.env.BROKER_URL);
    console.log("Conexão estabelecida com sucesso!");

    const channel = await connection.createChannel();
    console.log(`Declarando a fila ${queue}...`);
    await channel.assertQueue(queue, { durable: true });

    console.log(`Enviando mensagem para ${queue}: ${message}`);
    channel.sendToQueue(queue, Buffer.from(message), { persistent: true });
    console.log("Mensagem enviada com sucesso!");

    setTimeout(() => {
      channel.close();
      connection.close();
    }, 500);
  } catch (error) {
    console.error("Erro ao enviar mensagem:", error);
  }
}

module.exports = { sendMessage };
