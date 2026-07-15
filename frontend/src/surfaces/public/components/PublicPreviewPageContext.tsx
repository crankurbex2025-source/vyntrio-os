import { Link } from "react-router-dom";

export type PublicPreviewContextLink = {
  key: string;
  label: string;
  to: string;
};

export type PublicPreviewPageContextProps = {
  ariaLabel: string;
  links: PublicPreviewContextLink[];
  currentKey: string;
};

export function PublicPreviewPageContext({
  ariaLabel,
  links,
  currentKey,
}: PublicPreviewPageContextProps) {
  return (
    <nav className="vyn-public-preview-context" aria-label={ariaLabel}>
      <ol className="vyn-public-preview-context-list">
        {links.map((link) => (
          <li
            key={link.key}
            className={
              link.key === currentKey
                ? "vyn-public-preview-context-item vyn-public-preview-context-item-current"
                : "vyn-public-preview-context-item"
            }
          >
            {link.key === currentKey ? (
              <span aria-current="page">{link.label}</span>
            ) : (
              <Link to={link.to}>{link.label}</Link>
            )}
          </li>
        ))}
      </ol>
    </nav>
  );
}
