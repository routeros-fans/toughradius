package org.toughradius.entity;

import java.util.ArrayList;
import java.util.List;

public class MenuItem {

    private String id;
    private String icon;
    private String value;
    private List<MenuItem> data = new ArrayList<MenuItem>();

    public MenuItem(String id, String icon, String value) {
        this.id = id;
        this.icon = icon;
        this.value = value;
    }

    public void append(String id, String icon, String value)
    {
        this.data.add(new MenuItem(id,icon,value));
    }

    public void append(MenuItem item)
    {
        this.data.add(item);
    }

    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }

    public String getIcon() {
        return icon;
    }

    public void setIcon(String icon) {
        this.icon = icon;
    }

    public String getValue() {
        return value;
    }

    public void setValue(String value) {
        this.value = value;
    }

    public List<MenuItem> getData() {
        return data;
    }

    public void setData(List<MenuItem> data) {
        this.data = data;
    }
}
