import type { Metadata } from "next";
import HomeClient from "./home-client";

export const metadata: Metadata = {
  // absolute: a home não leva o sufixo "| Tô De Olho" do template
  title: { absolute: "Tô De Olho - Transparência no Senado" },
  alternates: { canonical: "/" },
};

export default function Page() {
  return <HomeClient />;
}
