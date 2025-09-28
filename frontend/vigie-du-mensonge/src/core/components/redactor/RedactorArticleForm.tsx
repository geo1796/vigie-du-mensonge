import * as React from "react";
import {z} from "zod";
import {useForm} from "react-hook-form";
import {zodResolver} from "@hookform/resolvers/zod";
import {useMutation, useQueryClient} from "@tanstack/react-query";
import {toast} from "@/core/utils/toast.ts";
import {Form, FormControl, FormField, FormItem, FormLabel, FormMessage} from "@/core/shadcn/components/ui/form.tsx";
import {Input} from "@/core/shadcn/components/ui/input.tsx";
import {Button} from "@/core/shadcn/components/ui/button.tsx";
import {ArticleCategories, ArticleCategoryLabels} from "@/core/models/articleCategory.ts";
import {useStore} from "@tanstack/react-store";
import {politiciansManager} from "@/core/dependencies/politician/politiciansManager.ts";
import {Select, SelectContent, SelectItem, SelectTrigger, SelectValue} from "@/core/shadcn/components/ui/select.tsx";
import {RedactorClient, type SaveRedactorArticle} from "@/core/dependencies/redactor/redactorClient.ts";
import type {Article} from "@/core/models/article.ts";
import {formatDateEN} from "@/core/utils/formatDate.ts";

export type RedactorArticleFormProps = {
    redactorClient: RedactorClient
    article?: Article
}

const publishSchema = z.object({
    title: z.string().min(1, "Titre requis").max(100, "100 caractères maximum").min(20, "20 caractères minimum"),
    eventDate: z.string().min(1, "Date de l'événement requise"), // we'll convert to Date on submit
    category: z.enum(Object.values(ArticleCategories)).nonoptional(),
    body: z.string().min(1, "Contenu requis").max(2000, "2000 caractères maximum").min(200, "200 caractères minimum"),
    tags: z.array(z.string().min(1).max(25)).min(1, "Au moins 1 tag").max(5, "5 tags maximum"),
    sources: z.array(z.url("URL invalide")).min(1, "Au moins 1 source").max(5, "5 sources maximum"),
    politicians: z.array(z.string()).max(5, "5 politiciens maximum"),
});

const draftSchema = z.object({
    title: z.string().min(1, "Titre requis").max(100, "100 caractères maximum").min(20, "20 caractères minimum"),
    eventDate: z.string().min(1, "Date de l'événement requise"), // we'll convert to Date on submit
    category: z.enum(Object.values(ArticleCategories)).nonoptional(),
    body: z.string().max(2000, "2000 caractères maximum"),
    tags: z.array(z.string().min(1).max(25)).max(5, "5 tags maximum"),
    sources: z.array(z.url("URL invalide")).max(5, "5 sources maximum"),
    politicians: z.array(z.string()).max(5, "5 politiciens maximum"),
});

const formSchema = z.discriminatedUnion("mode", [
    publishSchema.extend({mode: z.literal("publish")}),
    draftSchema.extend({mode: z.literal("draft")}),
]);

export type RedactorArticleFormInput = z.infer<typeof formSchema>;

function mapInput(input: RedactorArticleFormInput, articleId?: string): SaveRedactorArticle {
    return {
        id: articleId,
        title: input.title,
        body: input.body,
        eventDate: new Date(input.eventDate),
        tags: input.tags,
        politicianIds: input.politicians,
        sources: input.sources,
        category: input.category,
    };
}

export function RedactorArticleForm({redactorClient, article}: RedactorArticleFormProps) {
    const form = useForm<RedactorArticleFormInput>({
        resolver: zodResolver(formSchema),
        defaultValues: {
            mode: "draft",
            title: article?.title ?? "",
            eventDate: article?.eventDate ? formatDateEN(article.eventDate) : "",
            body: article?.body ?? "",
            category: article?.category ?? ArticleCategories.FALSEHOOD,
            tags: article?.tags ?? [],
            sources: article?.sources ?? [],
            politicians: article?.politicianIds ?? [],
        },
        mode: "onSubmit",
    });
    const queryClient = useQueryClient();
    const mutation = useMutation({
        mutationFn: async (input: RedactorArticleFormInput) => {
            return redactorClient.saveArticle(input.mode === "publish", mapInput(input, article?.id));
        },
        onSuccess:  () => {
            toast.success("Votre article a été enregistré.");
            if (article) {
                void queryClient.invalidateQueries({queryKey: ["redactor", "articles", article.reference]});
            }
            void queryClient.invalidateQueries({queryKey: ["redactor", "articles"]});
        },
        onError: () => {
            toast.error("Une erreur est survenue. Veuillez réessayer.");
        }
    });

    // Politicians search/select
    const allPoliticians = useStore(politiciansManager.politiciansStore);
    const [search, setSearch] = React.useState("");

    const selectedPoliticians = form.watch("politicians");

    const filtered = React.useMemo(() => {
        const q = search.trim().toLowerCase();
        if (!q) return allPoliticians.filter(p => !selectedPoliticians.includes(p.id));
        return allPoliticians.filter(p => !selectedPoliticians.includes(p.id) && p.fullName.toLowerCase().includes(q));
    }, [allPoliticians, search, selectedPoliticians]);

    function addPolitician(id: string) {
        const current = form.getValues("politicians");
        if (current.length >= 5) return;
        if (!current.includes(id)) form.setValue("politicians", [...current, id], {shouldValidate: true});
    }

    function removePolitician(id: string) {
        const current = form.getValues("politicians");
        form.setValue("politicians", current.filter(p => p !== id), {shouldValidate: true});
    }

    // Simple add input for tags and sources
    const [tagInput, setTagInput] = React.useState("");
    const [sourceInput, setSourceInput] = React.useState("");

    function addTag() {
        const v = tagInput.trim();
        if (!v) return;
        const tags = form.getValues("tags");
        if (tags.length >= 10) return;
        if (!tags.includes(v)) {
            form.setValue("tags", [...tags, v], {shouldValidate: true});
            setTagInput("");
        }
    }

    function removeTag(tag: string) {
        const tags = form.getValues("tags");
        form.setValue("tags", tags.filter(t => t !== tag), {shouldValidate: true});
    }

    function addSource() {
        const v = sourceInput.trim();
        if (!v) return;
        const sources = form.getValues("sources");
        if (sources.length >= 5) return;
        if (!sources.includes(v)) {
            form.setValue("sources", [...sources, v], {shouldValidate: true});
            setSourceInput("");
        }
    }

    function removeSource(u: string) {
        const sources = form.getValues("sources");
        form.setValue("sources", sources.filter(s => s !== u), {shouldValidate: true});
    }

    async function onSubmit(values: RedactorArticleFormInput) {
        await mutation.mutateAsync(values);
    }

    return (
        <div className="mx-auto w-full max-w-2xl">

            <Form {...form}>
                <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
                    <div className="space-y-1">
                        <h1 className="text-xl font-semibold">
                            {article ? "Modifier l'article" : "Créer un article"}
                        </h1>
                        <p className="text-sm text-muted-foreground">Renseignez les informations ci-dessous</p>
                    </div>


                    <FormField
                        control={form.control}
                        name="title"
                        render={({field}) => (
                            <FormItem>
                                <FormLabel>Titre</FormLabel>
                                <FormControl>
                                    <Input type="text" placeholder="Titre de l'article" {...field} />
                                </FormControl>
                                <FormMessage/>
                            </FormItem>
                        )}
                    />

                    <FormField
                        control={form.control}
                        name="eventDate"
                        render={({field}) => (
                            <FormItem>
                                <FormLabel>Date de l'événement</FormLabel>
                                <FormControl>
                                    <Input type="date" {...field} />
                                </FormControl>
                                <FormMessage/>
                            </FormItem>
                        )}
                    />

                    <FormField
                        control={form.control}
                        name="category"
                        render={({field}) => (
                            <FormItem>
                                <FormLabel>Catégorie</FormLabel>
                                <FormControl>
                                    <Select onValueChange={(selected) => {
                                        field.onChange(selected);
                                    }}>
                                        <SelectTrigger>
                                            <SelectValue placeholder={ArticleCategoryLabels[field.value]}/>
                                        </SelectTrigger>
                                        <SelectContent>
                                            {Object.entries(ArticleCategoryLabels).map(([value, label]) => (
                                                <SelectItem key={value} value={value}>{label}</SelectItem>
                                            ))}
                                        </SelectContent>
                                    </Select>
                                </FormControl>
                                <FormMessage/>
                            </FormItem>
                        )}
                    />

                    <FormField
                        control={form.control}
                        name="politicians"
                        render={() => (
                            <FormItem>
                                <FormLabel>Personnalités politiques impliquées. Si il/elle n'est pas présent(e) dans
                                    notre base de données, veuillez saisir un seul tag contenant à la fois son nom et
                                    prénom.</FormLabel>
                                <div className="space-y-2">
                                    <Input
                                        placeholder="Rechercher une personnalité politique"
                                        value={search}
                                        onChange={e => setSearch(e.target.value)}
                                    />
                                    <div className="max-h-40 overflow-auto rounded-md border">
                                        {filtered.length === 0 ? (
                                            <div className="p-2 text-sm text-muted-foreground">Aucun résultat</div>
                                        ) : (
                                            filtered.slice(0, 20).map(p => (
                                                <button
                                                    type="button"
                                                    key={p.id}
                                                    className="w-full text-left px-3 py-2 text-sm hover:bg-accent"
                                                    onClick={() => addPolitician(p.id)}
                                                    disabled={form.getValues("politicians").length >= 5}
                                                >
                                                    {p.fullName}
                                                </button>
                                            ))
                                        )}
                                    </div>
                                    <div className="flex flex-wrap gap-2">
                                        {selectedPoliticians.map(id => {
                                            const p = allPoliticians.find(pp => pp.id === id);
                                            if (!p) return null;
                                            return (
                                                <span key={id}
                                                      className="inline-flex items-center gap-2 rounded-md border px-2 py-1 text-xs">
                                                    {p.fullName}
                                                    <button type="button"
                                                            className="text-muted-foreground hover:text-foreground"
                                                            onClick={() => removePolitician(id)}>&times;</button>
                                                </span>
                                            );
                                        })}
                                    </div>
                                </div>
                                <FormMessage/>
                            </FormItem>
                        )}
                    />

                    <FormField
                        control={form.control}
                        name="body"
                        render={({field}) => (
                            <FormItem>
                                <FormLabel>Contenu</FormLabel>
                                <FormControl>
                                    <textarea
                                        className="border-input w-full min-h-32 rounded-md border bg-transparent px-3 py-2 text-sm"
                                        placeholder="Racontez les faits…"
                                        maxLength={2000}
                                        value={field.value}
                                        onChange={field.onChange}
                                    />
                                </FormControl>
                                <FormMessage/>
                            </FormItem>
                        )}
                    />

                    <FormField
                        control={form.control}
                        name="tags"
                        render={() => (
                            <FormItem>
                                <FormLabel>Tags</FormLabel>
                                <div className="flex gap-2">
                                    <Input
                                        placeholder="Ajouter un tag"
                                        value={tagInput}
                                        onChange={e => setTagInput(e.target.value)}
                                        onKeyDown={e => {
                                            if (e.key === 'Enter') {
                                                e.preventDefault();
                                                addTag();
                                            }
                                        }}
                                    />
                                    <Button type="button" onClick={addTag} variant="secondary">Ajouter</Button>
                                </div>
                                <div className="flex flex-wrap gap-2 pt-2">
                                    {form.getValues("tags").map(tag => (
                                        <span key={tag}
                                              className="inline-flex items-center gap-2 rounded-md border px-2 py-1 text-xs">
                                            #{tag}
                                            <button type="button"
                                                    className="text-muted-foreground hover:text-foreground"
                                                    onClick={() => removeTag(tag)}>&times;</button>
                                        </span>
                                    ))}
                                </div>
                                <FormMessage/>
                            </FormItem>
                        )}
                    />

                    <FormField
                        control={form.control}
                        name="sources"
                        render={() => (
                            <FormItem>
                                <FormLabel>Sources (URL)</FormLabel>
                                <div className="flex gap-2">
                                    <Input
                                        placeholder="https://exemple.com/article"
                                        value={sourceInput}
                                        onChange={e => setSourceInput(e.target.value)}
                                        onKeyDown={e => {
                                            if (e.key === 'Enter') {
                                                e.preventDefault();
                                                addSource();
                                            }
                                        }}
                                    />
                                    <Button type="button" onClick={addSource} variant="secondary">Ajouter</Button>
                                </div>
                                <div className="flex flex-col gap-2 pt-2">
                                    {form.getValues("sources").map(src => (
                                        <div key={src}
                                             className="flex items-center justify-between rounded-md border px-2 py-1 text-xs">
                                            <a href={src} target="_blank" rel="noreferrer"
                                               className="truncate max-w-40 sm:max-w-100 underline">
                                                {src}
                                            </a>
                                            <button type="button"
                                                    className="text-muted-foreground hover:text-foreground"
                                                    onClick={() => removeSource(src)}>&times;</button>
                                        </div>
                                    ))}
                                </div>
                                <FormMessage/>
                            </FormItem>
                        )}
                    />

                    <div className="flex flex-row gap-8 items-center justify-center">
                        <Button
                            type="submit"
                            disabled={form.formState.isSubmitting}
                            onClick={() => form.setValue("mode", "draft", {shouldValidate: false})}
                        >
                            Enregistrer
                        </Button>
                        {article &&
                            <Button
                                type="submit"
                                disabled={form.formState.isSubmitting}
                                onClick={() => form.setValue("mode", "publish", {shouldValidate: false})}
                            >
                                Publier
                            </Button>
                        }
                    </div>

                </form>
            </Form>
        </div>
    );
}