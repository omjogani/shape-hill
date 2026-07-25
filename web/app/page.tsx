import { HillEditor } from "@/components/organisms/hill-editor/HillEditor";

const slug = process.env.NEXT_PUBLIC_HILL_SLUG ?? "demo-billing";

export default function Home() {
  return <HillEditor slug={slug} />;
}
