export default function ToDo(props: any) {
  return (
    <div className="todo">
      <h3>{props.data.title}</h3>
      <p>{props.data.desc}</p>
    </div>
  );
}
