import type { DetailedHTMLProps, HTMLAttributes } from "react";
import type { CapWidget } from "cap-widget";

declare module "react" {
  namespace JSX {
    interface IntrinsicElements {
      "cap-widget": DetailedHTMLProps<HTMLAttributes<CapWidget>, CapWidget> & {
        required?: boolean;
        "data-cap-api-endpoint"?: string;
        "data-cap-hidden-field-name"?: string;
        "data-cap-worker-count"?: string;
        "data-cap-troubleshooting-url"?: string;
      };
    }
  }
}
