import {createFileRoute} from '@tanstack/react-router';
import {useQuery} from "@tanstack/react-query";
import {Spinner} from "@/core/shadcn/components/ui/spinner.tsx";
import {AdminUserProfile} from "@/core/components/admin/AdminUserProfile.tsx";

export const Route = createFileRoute('/admin/users/$userTag')({
    component: RouteComponent,
});

function RouteComponent() {
    const adminClient = Route.useRouteContext().adminClient;
    const userTag = Route.useParams().userTag;

    const {queryKey, queryFn} = adminClient.findUserByTag(userTag);

    const {data: user, isLoading, isError} = useQuery({
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

    if (!user) {
        return <div className="flex items-center justify-center h-screen">
            Une erreur est survenue. Veuillez réessayer.
        </div>;
    }

    return <div className="flex flex-col items-center gap-8 min-w-0 p-2">
        <AdminUserProfile user={user} adminClient={adminClient}/>
    </div>;
}
