import {
  EDITOR_TYPES,
  LANGUAGE_OPTIONS,
} from "@/components/EspressoConsole/constants";
import { cn } from "@/utils/index";
import { useRef, useState } from "react";
import useOnClickOutside from "@/hooks/useOnClickOutside";

const field = {
  id: "template_name",
  label: "Name",
  rules: {
    required: "This field is required",
  },
  className: "mb-6",
  defaultValue: "",
};

const EditorTab = ({
  formControls,
  setEditorType,
  editorType,
  editorProps,
}) => {
  const { register } = formControls;

  const [isDropdownOpen, setIsDropdownOpen] = useState(false);
  const [readOnlyEditorType, setReadOnlyEditorType] = useState(
    EDITOR_TYPES.GO_EDITOR
  );

  const copyRef = useRef(null);
  useOnClickOutside(copyRef, () => {
    setIsDropdownOpen(false);
  });

  const handleLanguageChange = (e, value) => {
    e.stopPropagation();
    setEditorType(value);
    setReadOnlyEditorType(value);
    setIsDropdownOpen(false);
  };

  return (
    <div className="absolute right-2 top-2 flex text-sm w-[90%]">
      <div className="flex w-full justify-between">
        <input
          autoFocus
          id={field.id}
          type="text"
          {...register(field.id, { required: field.rules?.required })}
          defaultValue={field.defaultValue || ""}
          placeholder={`Enter ${field.label.toLowerCase()}`}
          className="px-1 border-2 border-black rounded-md focus:outline-none"
        />
        <div className="flex border border-black rounded-md">
          <div
            className={cn("cursor-pointer p-1 px-2 ", {
              "rounded-l bg-black text-white":
                editorType === EDITOR_TYPES.HTML_EDITOR,
            })}
            onClick={() => setEditorType(EDITOR_TYPES.HTML_EDITOR)}
          >
            Live Editor
          </div>
          <div
            className={cn("cursor-pointer border-l border-black p-1 px-2 ", {
              "bg-black text-white": editorType === EDITOR_TYPES.JSON_EDITOR,
            })}
            onClick={() => setEditorType(EDITOR_TYPES.JSON_EDITOR)}
          >
            Mock values
          </div>
          <div
            className={cn(
              "relative flex w-32 cursor-pointer items-center justify-center rounded-r-md border-l border-black bg-[#FFDBDA] p-1 px-2 text-center  ",
              {
                "rounded-r bg-black text-white":
                  editorProps[editorType]?.options?.readOnly,
              }
            )}
            onClick={() => {
              setEditorType(readOnlyEditorType);

              setIsDropdownOpen(!isDropdownOpen);
            }}
            ref={copyRef}
          >
            Copy{" "}
            {LANGUAGE_OPTIONS.find(
              (option) => option.value === readOnlyEditorType
            )?.label || "Golang"}
            <div
              className={cn(
                "ml-1 flex  items-center justify-center transition-all",
                {
                  "rotate-180": isDropdownOpen,
                }
              )}
            >
              {/* <ZIcon unicode="E886" className=" text-xs" /> */}
            </div>
            <div
              className={cn(
                "absolute left-1/2 top-full size-4 -translate-x-1/2 rotate-45 border bg-white shadow-md",
                {
                  hidden: !isDropdownOpen,
                }
              )}
            />
            <div
              className={cn(
                "absolute right-0 top-full z-[99] mt-2 w-full  rounded-md border bg-white shadow-md ",
                {
                  hidden: !isDropdownOpen,
                }
              )}
            >
              <div className="relative">
                {LANGUAGE_OPTIONS.map((option, index) => (
                  <div
                    key={index}
                    className={cn(
                      "cursor-pointer  p-1 px-2 text-black first:rounded-t-md last:rounded-b-md ",
                      {
                        "bg-[#efefef]": readOnlyEditorType === option.value,
                        "hover:bg-[#efefef]":
                          readOnlyEditorType !== option.value,
                      }
                    )}
                    onClick={(e) => handleLanguageChange(e, option.value)}
                  >
                    {option.label}
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default EditorTab;
