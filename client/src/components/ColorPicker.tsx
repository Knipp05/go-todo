import { Fragment, useState } from "react";
import Wheel from "@uiw/react-color-wheel";
import { hsvaToHex } from "@uiw/color-convert";
import { Category } from "../App";

export default function ColorPicker(props: any) {
  const [hsva, setHsva] = useState({
    h: 0,
    s: 0,
    v: props.type === "color_header" ? 70 : 100,
    a: 1,
  });
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
