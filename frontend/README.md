# 🎨 Frontend - Tô De Olho

> **Plataforma de transparência política** com foco em **acessibilidade** e **mobile-first**

Este é um projeto [Next.js](https://nextjs.org) desenvolvido especificamente para o contexto brasileiro, priorizando acessibilidade WCAG 2.1 AA e experiência mobile-first.

## 📱 Mobile-First Approach

### **Contexto Brasileiro**
- **70% dos acessos via smartphone** (especialmente classes C/D/E)
- **Conectividade limitada**: 4G instável, franquia de dados
- **População alvo**: Adultos 35-65 anos, familiaridade média com tech
- **Dispositivos**: Android predominante, telas 5-6 polegadas

### **Princípios de Design**
- **Mobile-First**: Design começa em 375px, depois expande
- **Touch-Friendly**: Botões mínimo 44px x 44px
- **Typography**: Base 16px+ (evita zoom automático)
- **Performance**: Bundle <200KB, imagens WebP + lazy loading

### **Breakpoints Padrão**
```tsx
// ✅ Pattern Mobile-First obrigatório
<div className="
  grid grid-cols-1           // Mobile: 1 coluna
  md:grid-cols-2            // Tablet: 2 colunas  
  lg:grid-cols-3            // Desktop: 3 colunas
  gap-4 p-4                 // Mobile: padding menor
  md:gap-6 md:p-8           // Desktop: padding maior
">
```

## ♿ Acessibilidade (WCAG 2.1 AA)

- **Contraste**: Mínimo 4.5:1 em todos os elementos
- **Navegação**: Completa via teclado (Tab, Enter, Esc)
- **Screen readers**: Aria-labels e semantic HTML
- **Typography**: Textos legíveis, hierarquia clara

## 🚀 Getting Started

First, run the development server:

```bash
npm run dev
# or
yarn dev
# or
pnpm dev
# or
bun dev
```

Open [http://localhost:3000](http://localhost:3000) with your browser to see the result.

You can start editing the page by modifying `app/page.tsx`. The page auto-updates as you edit the file.

This project uses [`next/font`](https://nextjs.org/docs/app/building-your-application/optimizing/fonts) to automatically optimize and load [Geist](https://vercel.com/font), a new font family for Vercel.

## Learn More

To learn more about Next.js, take a look at the following resources:

- [Next.js Documentation](https://nextjs.org/docs) - learn about Next.js features and API.
- [Learn Next.js](https://nextjs.org/learn) - an interactive Next.js tutorial.

You can check out [the Next.js GitHub repository](https://github.com/vercel/next.js) - your feedback and contributions are welcome!

## Deploy on Vercel

The easiest way to deploy your Next.js app is to use the [Vercel Platform](https://vercel.com/new?utm_medium=default-template&filter=next.js&utm_source=create-next-app&utm_campaign=create-next-app-readme) from the creators of Next.js.

Check out our [Next.js deployment documentation](https://nextjs.org/docs/app/building-your-application/deploying) for more details.
