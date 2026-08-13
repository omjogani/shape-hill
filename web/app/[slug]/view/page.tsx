import { HillViewer } from "@/components/organisms/hill-viewer/HillViewer";

export default async function HillViewPage({ params }: PageProps<"/[slug]/view">) {
  const { slug } = await params;
  return <HillViewer slug={slug} />;
}
