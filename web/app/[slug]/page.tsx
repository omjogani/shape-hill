import { HillEditor } from "@/components/organisms/hill-editor/HillEditor";

export default async function HillPage({ params }: PageProps<"/[slug]">) {
  const { slug } = await params;
  return <HillEditor slug={slug} />;
}
