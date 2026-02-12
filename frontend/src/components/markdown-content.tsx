import ReactMarkdown from "react-markdown"
import rehypeHighlight from "rehype-highlight"
import remarkGfm from "remark-gfm"

import { cn } from "@/lib/utils"

interface MarkdownContentProps {
  content: string
  className?: string
}

export function MarkdownContent({ content, className }: MarkdownContentProps) {
  return (
    <div className={cn("markdown-body text-sm leading-6", className)}>
      <ReactMarkdown
        rehypePlugins={[rehypeHighlight]}
        remarkPlugins={[remarkGfm]}
        components={{
          p: ({ className, ...props }) => <p className={cn("mb-3 last:mb-0", className)} {...props} />,
          h1: ({ className, ...props }) => <h1 className={cn("mb-3 mt-4 text-xl font-semibold first:mt-0", className)} {...props} />,
          h2: ({ className, ...props }) => <h2 className={cn("mb-2 mt-4 text-lg font-semibold first:mt-0", className)} {...props} />,
          h3: ({ className, ...props }) => <h3 className={cn("mb-2 mt-3 text-base font-semibold first:mt-0", className)} {...props} />,
          ul: ({ className, ...props }) => <ul className={cn("mb-3 list-disc pl-5", className)} {...props} />,
          ol: ({ className, ...props }) => <ol className={cn("mb-3 list-decimal pl-5", className)} {...props} />,
          li: ({ className, ...props }) => <li className={cn("mb-1", className)} {...props} />,
          blockquote: ({ className, ...props }) => (
            <blockquote className={cn("mb-3 border-l-2 border-border pl-3 italic text-muted-foreground", className)} {...props} />
          ),
          code: ({ className, children, ...props }) => {
            const isInline = !String(className ?? "").includes("language-")
            if (isInline) {
              return (
                <code className="rounded bg-muted px-1 py-0.5 font-mono text-xs" {...props}>
                  {children}
                </code>
              )
            }
            return (
              <code className={cn("block overflow-x-auto rounded-md bg-muted p-3 font-mono text-xs", className)} {...props}>
                {children}
              </code>
            )
          },
          pre: ({ className, ...props }) => <pre className={cn("mb-3 overflow-x-auto", className)} {...props} />,
          table: ({ className, ...props }) => <table className={cn("mb-3 w-full border-collapse text-xs", className)} {...props} />,
          th: ({ className, ...props }) => <th className={cn("border border-border bg-muted px-2 py-1 text-left", className)} {...props} />,
          td: ({ className, ...props }) => <td className={cn("border border-border px-2 py-1", className)} {...props} />,
          a: ({ className, ...props }) => <a className={cn("text-primary underline underline-offset-2", className)} target="_blank" rel="noreferrer" {...props} />,
        }}
      >
        {content}
      </ReactMarkdown>
    </div>
  )
}
