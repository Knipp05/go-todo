import { Fragment, useEffect, useState } from "react";
import Wheel from "@uiw/react-color-wheel";
import { hsvaToHex, hexToHsva } from "@uiw/color-convert";
import { Category } from "../App";

export default function ColorPicker(props: any) {
  const [hsva, setHsva] = useState({
    h: 0,
    s: 0,
    v: props.type === "color_header" ? 70 : 100,
    a: 1,
  });
  useEffect(() => {
    if (props.dialogType === "edit") {
      setHsva(hexToHsva(props.categoryColor));
    } else if (props.dialogType === "create") {
      setHsva({
        h: 0,
        s: 0,
        v: props.type === "color_header" ? 70 : 100,
        a: 1,
      });
    }
  }, [props.category, props.dialogType]);
  return (
    <>
      <Fragment>
        <h3>
          {props.type === "color_header"
            ? "Farbe für den Aufgabenheader wählen"
            : "Farbe für den Aufgabenbody wählen"}
        </h3>
        <Wheel
          color={hsva}
          onChange={(color) => {
            props.setCategory((oldCategory: Category) => {
              return { ...oldCategory, [props.type]: hsvaToHex(hsva) };
            });
            setHsva({ ...hsva, ...color.hsva });
          }}
        />
      </Fragment>
    </>
  );
}
