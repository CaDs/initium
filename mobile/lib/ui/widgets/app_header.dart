// Wonderize: minimalist app header with back button + title + actions.

import 'package:flutter/material.dart';

import '../app_scaffold.dart';
import 'circle_btn.dart';

class AppHeader extends StatelessWidget implements PreferredSizeWidget {
  const AppHeader({
    super.key,
    this.title,
    this.leading,
    this.trailing,
    this.showBackBtn = true,
    this.backgroundColor,
    this.height = 64,
    this.titleStyle,
  });

  final String? title;
  final Widget? leading;
  final List<Widget>? trailing;
  final bool showBackBtn;
  final Color? backgroundColor;
  final double height;
  final TextStyle? titleStyle;

  @override
  Size get preferredSize => Size.fromHeight(height);

  @override
  Widget build(BuildContext context) {
    final lead = leading ??
        (showBackBtn && Navigator.of(context).canPop() ? const BackBtn() : null);

    return Material(
      color: backgroundColor ?? Colors.transparent,
      elevation: 0,
      child: SafeArea(
        bottom: false,
        child: SizedBox(
          height: height,
          child: Padding(
            padding: EdgeInsets.symmetric(horizontal: $styles.insets.sm),
            child: Stack(
              children: [
                if (lead != null)
                  Align(alignment: Alignment.centerLeft, child: lead),
                if (title != null)
                  Center(
                    child: Text(
                      title!,
                      style: titleStyle ??
                          $styles.text.title1
                              .copyWith(color: $styles.colors.fg),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                if (trailing != null && trailing!.isNotEmpty)
                  Align(
                    alignment: Alignment.centerRight,
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        for (final t in trailing!)
                          Padding(
                            padding: EdgeInsets.only(left: $styles.insets.xs),
                            child: t,
                          ),
                      ],
                    ),
                  ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
