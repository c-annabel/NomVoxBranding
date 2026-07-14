import IntakeForm from "@/components/IntakeForm";
import type { IntakePayload } from "@/lib/types";

// Temporary client wrapper — page becomes interactive in ST-03 when
// name cards and the generate API are wired up.
import HomeClient from "./HomeClient";

export default function Home() {
  return <HomeClient />;
}
