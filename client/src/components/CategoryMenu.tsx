import {
  Menu,
  Divider,
  MenuItem,
  Dialog,
  DialogTitle,
  TextField,
  DialogActions,
  Button,
  DialogContentText,
  DialogContent,
  Grid,
  IconButton,
} from "@mui/material";
import DeleteIcon from "@mui/icons-material/Delete";
import EditIcon from "@mui/icons-material/Edit";
import { BASE_URL, Category } from "../App";
import { useState } from "react";
import ColorPicker from "./ColorPicker";

export default function CategoryMenu(props: any) {
  const [category, setCategory] = useState<Category>({
    cat_name: "",
    color_header: "#00a4ba",
    color_body: "#00ceea",
  });
  const [showDialog, setShowDialog] = useState(false);

  function handleInput(event: any) {
    setCategory((oldCategory) => {
      return { ...oldCategory, [event.target.name]: event.target.value };
    });
  }

  const submitInput = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const token = sessionStorage.getItem("token");
    if (token && category.cat_name !== "default") {
      try {
        const res = await fetch(BASE_URL + `/${props.user.name}/categories`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            Authorization: token,
          },
          body: JSON.stringify(category),
        });

        const data = await res.json();

        if (!res.ok) {
          throw new Error(data.error || "Unbekannter Fehler aufgetreten");
        }
        props.setUser((oldUser: any) => {
          var newCategories = [
            ...oldUser.categories,
            {
              cat_name: category.cat_name,
              color_header: category.color_header,
              color_body: category.color_body,
            },
          ];
          return { ...oldUser, categories: newCategories };
        });
        setShowDialog(false);
        setCategory({
          cat_name: "",
          color_header: "#00a4ba",
          color_body: "#00ceea",
        });
      } catch (error: any) {
        throw new Error("Fehler beim Erstellen/Ändern der Aufgabe aufgetreten");
      }
    }
  };

  function editCategory() {
    //TODO!
  }

  function deleteCategory() {
    //TODO!
  }

  const categoryElements = props.categories.map(
    (category: Category, idx: number) => (
      <MenuItem
        key={idx}
        sx={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
        }}
        onClick={() =>
          props.changeCategory(
            category.cat_name,
            category.color_header,
            category.color_body
          )
        }
      >
        {category.cat_name !== "default"
          ? category.cat_name
          : "nicht kategorisiert"}
        <div>
          {category.cat_name !== "default" && (
            <IconButton size="small" onClick={editCategory}>
              <EditIcon fontSize="small" />
            </IconButton>
          )}
          {category.cat_name !== "default" && (
            <IconButton size="small" onClick={deleteCategory}>
              <DeleteIcon fontSize="small" />
            </IconButton>
          )}
        </div>
      </MenuItem>
    )
  );

  return (
    <>
      <Menu
        id="basic-menu"
        anchorEl={props.anchorEl}
        open={props.open}
        onClose={props.handleClose}
        MenuListProps={{
          "aria-labelledby": "basic-button",
        }}
      >
        {categoryElements}
        <Divider sx={{ my: 0.5 }} />
        <MenuItem onClick={() => setShowDialog(true)}>
          Kategorie hinzufügen
        </MenuItem>
      </Menu>
      <Dialog
        open={showDialog}
        onClose={() => setShowDialog(false)}
        PaperProps={{
          component: "form",
          onSubmit: (event: React.FormEvent<HTMLFormElement>) =>
            submitInput(event),
        }}
      >
        <DialogTitle>Neue Kategorie anlegen</DialogTitle>
        <DialogContent>
          <DialogContentText>
            Bitte Kategorienamen und optional eine Farbe angeben
          </DialogContentText>
          <TextField
            required
            id="cat_name"
            name="cat_name"
            label="Name"
            type="text"
            variant="standard"
            value={category.cat_name}
            onChange={handleInput}
            fullWidth
            margin="normal"
          />
          <Grid container spacing={2}>
            <Grid item xs={6}>
              <ColorPicker setCategory={setCategory} type="color_header" />
            </Grid>
            <Grid item xs={6}>
              <ColorPicker setCategory={setCategory} type="color_body" />
            </Grid>
          </Grid>
        </DialogContent>
        <DialogActions>
          <Button type="submit">Kategorie anlegen</Button>
        </DialogActions>
      </Dialog>
    </>
  );
}
