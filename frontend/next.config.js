/** @type {import('next').NextConfig} */
const nextConfig = {
  // Configurações para Docker
  output: 'standalone',
  
  // Configurações de imagem
  images: {
    domains: ['www.camara.leg.br'],
    unoptimized: true
  },
  
  // Configurações de ambiente
  env: {
    NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'
  }
}

module.exports = nextConfig
