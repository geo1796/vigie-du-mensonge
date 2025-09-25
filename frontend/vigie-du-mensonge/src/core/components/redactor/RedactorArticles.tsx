import {useStore} from "@tanstack/react-store";
import {ArticleCardWrap} from "@/core/components/article/ArticleCardWrap.tsx";
import {Link} from "@/core/utils/router.ts";
import {redactorArticlesManager} from "@/core/dependencies/redactor/redactorArticlesManager.ts";
import {SquarePen} from "lucide-react";

export function RedactorArticles() {
    const articles = useStore(redactorArticlesManager.articlesStore);
    const err = useStore(redactorArticlesManager.errStore);

    if (err) {
        return (
            <div className="flex items-center justify-center h-screen">
                Une erreur est survenue. Veuillez réessayer.
            </div>
        );
    }

    return (
        <div className="flex flex-col items-center gap-8 min-w-0 py-2">
            <Link
                to="/redactor/article-form"
                search={{articleID: undefined}}
                className="inline-flex items-center rounded-md border px-3 py-2 text-sm font-medium hover:bg-accent"
            >
                Ajouter un article
            </Link>
            <ArticleCardWrap articles={articles}
                             articleNavButton={(article) => RedactorArticleNavButton({article})}
            />
        </div>
    );
}

function RedactorArticleNavButton(props: { article: { id: string } }) {
    return (
        <Link
            to="/redactor/article-form"
            search={{articleID: props.article.id}}
            className="inline-flex items-center"
        >
            <SquarePen/>
        </Link>
    );
}