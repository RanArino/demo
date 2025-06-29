
export default function SpacesLayout({
  children,
  modal,
  miniMapModal,
}: {
  children: React.ReactNode;
  modal: React.ReactNode;
  miniMapModal: React.ReactNode;
}) {
  return (
    <div>
      {children}
      {modal}
      {miniMapModal}
    </div>
  );
}
