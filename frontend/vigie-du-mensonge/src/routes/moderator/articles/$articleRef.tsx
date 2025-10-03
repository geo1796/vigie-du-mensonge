import {createFileRoute} from '@tanstack/react-router';
import {useQuery} from "@tanstack/react-query";
import {ModeratorArticlesByReference} from "@/core/components/moderator/ModeratorArticlesByReference.tsx";
import {Spinner} from "@/core/shadcn/components/ui/spinner.tsx";

export const Route = createFileRoute('/moderator/articles/$articleRef')({
    component: RouteComponent,
});

function RouteComponent() {
    const {articleRef} = Route.useParams();
    const moderatorClient = Route.useRouteContext().moderatorClient;

    const {queryKey, queryFn} = moderatorClient.findModeratorArticlesByRef(articleRef);

    const {data: articles, isLoading, isError} = useQuery({
        queryKey,
        queryFn,
    });

    if (isError) {
        return <div className="flex items-center justify-center h-screen">
            Une erreur est survenue. Veuillez réessayer.
        </div>;
    }

    if (isLoading) {
        return (
            <div className="flex flex-col gap-2 items-center justify-center h-screen">
                Chargement en cours...
                <Spinner/>
            </div>
        );
    }

    if (!articles || articles.length === 0) {
        return (
            <div className="flex flex-col items-center justify-center h-screen">
                Une erreur est survenue. Veuillez réessayer.
            </div>
        );
    }

    return <ModeratorArticlesByReference articles={articles} moderatorClient={moderatorClient}/>;
}
