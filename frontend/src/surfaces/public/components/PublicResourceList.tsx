import { Link } from "react-router-dom";

export type PublicResourceItem = {
  title: string;
  description: string;
  statusLabel?: string;
  to?: string;
};

export type PublicResourceListProps = {
  items: PublicResourceItem[];
  headingId?: string;
};

export function PublicResourceList({ items, headingId }: PublicResourceListProps) {
  return (
    <ul className="vyn-public-resource-list" aria-labelledby={headingId}>
      {items.map((item) => (
        <li key={item.title} className="vyn-public-resource-item">
          <div className="vyn-public-resource-item-header">
            {item.to ? (
              <Link className="vyn-public-resource-title" to={item.to}>
                {item.title}
              </Link>
            ) : (
              <span className="vyn-public-resource-title">{item.title}</span>
            )}
            {item.statusLabel ? (
              <span className="vyn-public-resource-status">{item.statusLabel}</span>
            ) : null}
          </div>
          <p className="vyn-public-resource-description">{item.description}</p>
        </li>
      ))}
    </ul>
  );
}
