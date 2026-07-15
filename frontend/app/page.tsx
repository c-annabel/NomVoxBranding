import HomeClient from "./HomeClient";
import ErrorBoundary from "@/components/ErrorBoundary";

export default function Home() {
  return (
    <ErrorBoundary>
      <HomeClient />
    </ErrorBoundary>
  );
}
