type PageHeaderProps = {
  eyebrow?: string;
  title: string;
  description?: string;
};

export function PageHeader({ eyebrow, title, description }: PageHeaderProps) {
  return (
    <header className="page-header">
      {eyebrow ? <p className="page-header-eyebrow">{eyebrow}</p> : null}
      <h1 className="page-header-title">{title}</h1>
      {description ? <p className="page-header-description">{description}</p> : null}
    </header>
  );
}
