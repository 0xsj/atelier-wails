# forms

Thin frontend form controls. `Input`, `Textarea`, `SearchField`, `InputGroup`,
and `FieldMessage` cover native entry and validation presentation; `Select`,
`Combobox`, `Slider`, `Checkbox`, `Switch` and `RadioGroup` wrap Bits UI behavior
while Atelier owns the visual layer. `Field` wires a label, optional description
and message to any control through generated ids so the control is named and
described without the consumer managing attributes. Consumers own values,
filtering, validation, and submission.
