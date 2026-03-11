#!/usr/bin/env python3
"""
Library Organization Script
Processes LibIB CSV exports to generate HTML organization guide with space calculations
"""

import csv
import html
from pathlib import Path
from typing import Dict, List, Tuple, Optional

# Physical shelf dimensions (inches)
SHELF_DIMENSIONS = {
    'office_a': {'rows': 5, 'width': 32, 'depth': 26},  # Office shelf A
    'office_b': {'rows': 4, 'width': 32, 'depth': 26},  # Office shelf B  
    'dining': {'rows': 3, 'width': 24, 'depth': 24},    # Dining room
    'hallway': {'rows': 2, 'width': 16, 'depth': 16},   # Hallway
    'crate_h': {'rows': 1, 'width': 32, 'depth': 12},   # Crate horizontal
    'crate_v': {'rows': 1, 'width': 12, 'depth': 32}    # Crate vertical
}

# Shelf organization categories (rebalanced)
ORGANIZATION = {
    'office_a': {
        1: "Core Language Learning",
        2: "Essential Programming & Math", 
        3: "Core Science Textbooks",
        4: "Essential Philosophy",
        5: "Professional Writing & Development"
    },
    'office_b': {
        1: "Philosophy & Political Thought",
        2: "Academic Textbooks & References", 
        3: "Essays & Collections",
        4: "Gaming & Strategy"
    },
    'dining': {
        1: "Japanese Language & Asian Cooking",
        2: "Chinese Language & Global Cooking",
        3: "Other Languages & Kitchen References"
    },
    'hallway': {
        1: "History & Historical Works",
        2: "Literature, Fiction & Classics"
    },
    'crate_h': {
        1: "General Overflow Storage"
    },
    'crate_v': {
        1: "Mathematics, Science & References"
    }
}

def estimate_book_width(pages: Optional[int]) -> float:
    """Estimate book width in inches based on page count"""
    if not pages or pages <= 0:
        return 1.0  # Default width for unknown
    
    # Typical book spine width calculation
    if pages <= 100:
        return 0.5
    elif pages <= 200:
        return 0.75
    elif pages <= 300:
        return 1.0
    elif pages <= 500:
        return 1.25
    elif pages <= 750:
        return 1.5
    else:
        return 2.0

def categorize_book(title: str, tags: str, group: str, description: str, item: Dict = None) -> Tuple[str, int]:
    """Categorize a book into shelf and row - rebalanced for actual capacity"""
    title_lower = title.lower()
    tags_lower = tags.lower() if tags else ""
    group_lower = group.lower() if group else ""
    desc_lower = description.lower() if description else ""
    
    combined = f"{title_lower} {tags_lower} {group_lower} {desc_lower}"
    
    # PRIORITY: Core language learning (keep on Office A)
    if any(word in combined for word in ['genki', 'tobira', 'integrated chinese', 'kanji for international']):
        return 'office_a', 1
    
    # PRIORITY: Essential math/CS references (keep on Office A) 
    if any(word in combined for word in ['algorithms', 'clean code', 'programming language', 'knuth', 'art of computer']):
        return 'office_a', 2
        
    # PRIORITY: Core textbooks (keep on Office A)
    if any(word in combined for word in ['physics', 'chemistry']) and 'textbook' in combined:
        return 'office_a', 3
        
    # PRIORITY: Essential philosophy (keep on Office A)
    if any(word in combined for word in ['kant', 'groundwork', 'nietzsche', 'republic', 'analects']):
        return 'office_a', 4
        
    # PRIORITY: Professional development (keep on Office A)
    if any(word in combined for word in ['writer\'s reference', 'writing', 'professional']):
        return 'office_a', 5
    
    # MOVE TO DINING: All other language materials (spread the load)
    if any(word in combined for word in ['japanese', 'chinese', 'korean', 'language', 'dictionary', 'grammar']) and 'cooking' not in combined:
        if 'japanese' in combined or 'kanji' in combined:
            return 'dining', 1  # Japanese language overflow
        elif 'chinese' in combined:
            return 'dining', 2  # Chinese language overflow  
        else:
            return 'dining', 3  # Other languages
    
    # MOVE TO HALLWAY: History and literature (utilize hallway capacity)
    if any(word in combined for word in ['history', 'historical', 'revolution', 'war', 'ancient', 'medieval']):
        return 'hallway', 1
        
    if any(word in combined for word in ['dover', 'classics', 'poetry', 'shakespeare', 'literature', 'fiction', 'novel']):
        return 'hallway', 2
    
    # MOVE TO CRATE_V: Math, science, references (utilize vertical crate)
    if any(word in combined for word in ['mathematics', 'math', 'dover', 'logic', 'science', 'reference']):
        return 'crate_v', 1
        
    # Keep original office B for specialized items
    if any(word in combined for word in ['game', 'gaming', 'rpg', 'd&d', 'strategy', 'chess']):
        return 'office_b', 4
        
    # Philosophy overflow to Office B
    if any(word in combined for word in ['philosophy', 'political', 'ethics', 'tao']):
        return 'office_b', 1
        
    # Academic references to Office B
    if any(word in combined for word in ['textbook', 'academic', 'university']):
        return 'office_b', 2
        
    # Essays and collections to Office B  
    if any(word in combined for word in ['uncle john', 'essay', 'anthology', 'collection']):
        return 'office_b', 3
    
    # Keep cooking in dining room (original plan works)
    if any(word in combined for word in ['cooking', 'cook', 'food', 'recipe', 'kitchen', 'culinary']):
        if any(word in combined for word in ['asian', 'japanese', 'chinese', 'sushi', 'wok']):
            return 'dining', 1
        elif any(word in combined for word in ['bittman', 'international', 'global']):
            return 'dining', 2
        else:
            return 'dining', 3
    
    # Default to crate horizontal for remaining items
    return 'crate_h', 1

def parse_csv_file(filepath: Path) -> List[Dict]:
    """Parse a CSV file, handling multiline descriptions properly"""
    items = []
    
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            # Use csv.QUOTE_ALL to handle embedded quotes and newlines
            reader = csv.DictReader(f, quoting=csv.QUOTE_MINIMAL)
            
            for row in reader:
                if row.get('item_type') in ['book', 'videogame', 'music']:
                    items.append(row)
                    
    except Exception as e:
        print(f"Error reading {filepath}: {e}")
        
    return items

def calculate_space_usage(items: List[Dict], categorizer_func=None) -> Dict:
    """Calculate space usage for each shelf"""
    if categorizer_func is None:
        categorizer_func = categorize_book
        
    shelf_usage = {}
    
    for shelf_id in SHELF_DIMENSIONS:
        shelf_usage[shelf_id] = {
            'rows': {},
            'total_width': 0,
            'available_width': SHELF_DIMENSIONS[shelf_id]['width'] * SHELF_DIMENSIONS[shelf_id]['rows'],
            'utilization': 0
        }
        
        for row_num in range(1, SHELF_DIMENSIONS[shelf_id]['rows'] + 1):
            shelf_usage[shelf_id]['rows'][row_num] = {'width': 0, 'items': []}
    
    for item in items:
        if item['item_type'] == 'book':
            shelf_id, row_num = categorizer_func(
                item.get('title', ''),
                item.get('tags', ''),
                item.get('group', ''),
                item.get('description', ''),
                item  # Pass full item for adaptive logic
            )
            
            pages = 0
            try:
                pages = int(item.get('length', 0) or 0)
            except (ValueError, TypeError):
                pages = 0
                
            width = estimate_book_width(pages)
            
            shelf_usage[shelf_id]['rows'][row_num]['width'] += width
            shelf_usage[shelf_id]['rows'][row_num]['items'].append({
                'title': item.get('title', 'Unknown'),
                'author': item.get('creators', 'Unknown'),
                'pages': pages,
                'width': width,
                'item': item
            })
    
    # Calculate total usage
    for shelf_id in shelf_usage:
        total_width = sum(row['width'] for row in shelf_usage[shelf_id]['rows'].values())
        shelf_usage[shelf_id]['total_width'] = total_width
        shelf_usage[shelf_id]['utilization'] = (total_width / shelf_usage[shelf_id]['available_width']) * 100
    
    return shelf_usage

def create_adaptive_categorizer(utilization_data: Dict, iteration: int):
    """Create an adaptive categorizer that adjusts based on utilization"""
    
    def adaptive_categorize_book(title: str, tags: str, group: str, description: str, item: Dict) -> Tuple[str, int]:
        title_lower = title.lower()
        tags_lower = tags.lower() if tags else ""
        group_lower = group.lower() if group else ""
        desc_lower = description.lower() if description else ""
        
        combined = f"{title_lower} {tags_lower} {group_lower} {desc_lower}"
        
        # Helper function to find best available space
        def find_best_shelf(preferred_shelves, fallback_shelves=None):
            for shelf_id, row_num in preferred_shelves:
                if utilization_data.get(shelf_id, {}).get('utilization', 0) < 85:
                    return shelf_id, row_num
            
            if fallback_shelves:
                for shelf_id, row_num in fallback_shelves:
                    if utilization_data.get(shelf_id, {}).get('utilization', 0) < 95:
                        return shelf_id, row_num
            
            # Last resort - find least utilized
            min_util = float('inf')
            best_option = ('office_b', 1)
            for shelf_id in SHELF_DIMENSIONS:
                util = utilization_data.get(shelf_id, {}).get('utilization', 0)
                if util < min_util:
                    min_util = util
                    best_option = (shelf_id, 1)
            
            return best_option
        
        # TIER 1: Absolutely essential items (always office A if possible)
        if any(word in combined for word in ['genki i', 'genki ii', 'tobira', 'clean code', 'algorithms']):
            return find_best_shelf([('office_a', 1), ('office_a', 2)], [('office_b', 1)])
        
        # TIER 2: Core language learning
        if any(word in combined for word in ['japanese', 'kanji']) and any(word in combined for word in ['textbook', 'dictionary', 'learning']):
            return find_best_shelf([('office_a', 1), ('dining', 1)], [('office_b', 3)])
            
        if any(word in combined for word in ['chinese']) and any(word in combined for word in ['textbook', 'integrated', 'learning']):
            return find_best_shelf([('office_a', 1), ('dining', 2)], [('office_b', 3)])
        
        # TIER 3: Programming and math
        if any(word in combined for word in ['programming', 'computer', 'software']):
            return find_best_shelf([('office_a', 2), ('office_b', 2)], [('crate_v', 1)])
            
        if any(word in combined for word in ['mathematics', 'math', 'algorithms']) and 'game' not in combined:
            return find_best_shelf([('office_a', 2), ('office_b', 2), ('crate_v', 1)])
        
        # TIER 4: Science and textbooks
        if any(word in combined for word in ['physics', 'chemistry', 'science']) and 'cooking' not in combined:
            return find_best_shelf([('office_a', 3), ('office_b', 2), ('crate_v', 1)])
        
        # TIER 5: Philosophy 
        if any(word in combined for word in ['philosophy', 'kant', 'nietzsche', 'plato']):
            return find_best_shelf([('office_a', 4), ('office_b', 1)])
        
        # TIER 6: History
        if any(word in combined for word in ['history', 'historical', 'war', 'revolution']):
            return find_best_shelf([('office_b', 1), ('hallway', 1)], [('crate_h', 1)])
        
        # TIER 7: Literature and fiction
        if any(word in combined for word in ['literature', 'fiction', 'novel', 'poetry', 'classics']):
            return find_best_shelf([('office_b', 3), ('hallway', 2)], [('crate_h', 1)])
        
        # TIER 8: Languages (overflow)
        if any(word in combined for word in ['language', 'dictionary', 'grammar']) and 'cooking' not in combined:
            return find_best_shelf([('dining', 3), ('office_b', 3)], [('crate_v', 1)])
        
        # TIER 9: Gaming
        if any(word in combined for word in ['game', 'gaming', 'rpg', 'chess', 'strategy']):
            return find_best_shelf([('office_b', 4)], [('crate_h', 1)])
        
        # TIER 10: Cooking (stay in dining if possible)
        if any(word in combined for word in ['cooking', 'cook', 'food', 'recipe', 'kitchen']):
            if any(word in combined for word in ['asian', 'japanese', 'chinese', 'sushi']):
                return find_best_shelf([('dining', 1)], [('dining', 3)])
            else:
                return find_best_shelf([('dining', 2), ('dining', 3)])
        
        # TIER 11: Essays and collections
        if any(word in combined for word in ['uncle john', 'essay', 'anthology', 'collection']):
            return find_best_shelf([('office_b', 3), ('hallway', 2)], [('crate_h', 1)])
        
        # TIER 12: Everything else
        return find_best_shelf([('office_b', 1), ('office_b', 2), ('office_b', 3)], [('crate_h', 1)])
    
    return adaptive_categorize_book

def balanced_categorize_book(title: str, tags: str, group: str, description: str, item: Dict = None) -> Tuple[str, int]:
    """Simple balanced categorizer that considers capacity constraints"""
    title_lower = title.lower()
    tags_lower = tags.lower() if tags else ""
    group_lower = group.lower() if group else ""
    desc_lower = description.lower() if description else ""
    
    combined = f"{title_lower} {tags_lower} {group_lower} {desc_lower}"
    
    # Calculate total collection size to determine distribution strategy
    # Office A (160" capacity) - Most essential items only
    if any(word in combined for word in ['genki i', 'genki ii', 'tobira', 'clean code']):
        return 'office_a', 1
    if any(word in combined for word in ['algorithms', 'art of computer', 'programming language']):
        return 'office_a', 2
    if any(word in combined for word in ['physics', 'chemistry']) and 'textbook' in combined:
        return 'office_a', 3
    if any(word in combined for word in ['kant', 'groundwork', 'republic']):
        return 'office_a', 4
    if any(word in combined for word in ['writer\'s reference', 'writing']):
        return 'office_a', 5
        
    # Office B (128" capacity) - Secondary academic materials
    if any(word in combined for word in ['programming', 'computer', 'software', 'math']):
        return 'office_b', 1
    if any(word in combined for word in ['philosophy', 'political', 'ethics']):
        return 'office_b', 2
    if any(word in combined for word in ['reference', 'textbook', 'academic']):
        return 'office_b', 3
    if any(word in combined for word in ['game', 'gaming', 'rpg', 'chess', 'strategy']):
        return 'office_b', 4
        
    # Dining room (72" capacity) - Language + cooking
    if any(word in combined for word in ['cooking', 'cook', 'food', 'recipe', 'kitchen']):
        return 'dining', 1
    if any(word in combined for word in ['japanese', 'kanji']) and 'cooking' not in combined:
        return 'dining', 2  
    if any(word in combined for word in ['chinese', 'korean', 'language']) and 'cooking' not in combined:
        return 'dining', 3
        
    # Hallway (32" capacity) - Light reading
    if any(word in combined for word in ['fiction', 'novel', 'poetry']):
        return 'hallway', 1
    if any(word in combined for word in ['uncle john', 'essay', 'anthology']):
        return 'hallway', 2
        
    # Crate H (32" capacity) - History 
    if any(word in combined for word in ['history', 'historical', 'war', 'revolution']):
        return 'crate_h', 1
        
    # Crate V (12" capacity) - Small references
    if any(word in combined for word in ['dover', 'classics', 'literature']) and 'japanese' not in combined:
        return 'crate_v', 1
    
    # Default overflow to Office B (has most remaining capacity)
    return 'office_b', 2

def iterative_balance(items: List[Dict], max_iterations: int = 10) -> Tuple[Dict, int]:
    """Simple balanced approach - just use the balanced categorizer"""
    print(f"🔄 Using balanced categorization approach...")
    
    # Use the balanced categorizer
    space_usage = calculate_space_usage(items, balanced_categorize_book)
    
    print(f"📊 Utilization with balanced approach:")
    over_capacity = []
    for shelf_id, data in space_usage.items():
        util = data['utilization']
        print(f"   {shelf_id}: {util:.1f}% utilization")
        if util > 100:
            over_capacity.append((shelf_id, util))
    
    if over_capacity:
        print(f"⚠️  {len(over_capacity)} shelves still over capacity")
        print("💡 Consider: 1) Getting additional shelving, 2) Digital copies, 3) Storage rotation")
    else:
        print("✅ All shelves within capacity!")
    
    return space_usage, 1

def generate_html(all_items: List[Dict], space_usage: Dict) -> str:
    """Generate HTML organization guide"""
    
    html_content = f"""
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Library Organization Guide</title>
    <style>
        body {{ font-family: Arial, sans-serif; margin: 20px; background-color: #f5f5f5; }}
        .container {{ max-width: 1200px; margin: 0 auto; background: white; padding: 20px; border-radius: 8px; }}
        h1 {{ color: #2c3e50; text-align: center; border-bottom: 3px solid #3498db; padding-bottom: 10px; }}
        h2 {{ color: #34495e; border-left: 4px solid #3498db; padding-left: 10px; }}
        .summary {{ background: #ecf0f1; padding: 15px; border-radius: 5px; margin-bottom: 20px; }}
        .shelf {{ margin-bottom: 30px; border: 1px solid #ddd; border-radius: 5px; overflow: hidden; }}
        .shelf-header {{ background: #3498db; color: white; padding: 10px; font-weight: bold; }}
        .row {{ border-bottom: 1px solid #eee; }}
        .row-header {{ background: #ecf0f1; padding: 8px; font-weight: bold; display: flex; justify-content: space-between; }}
        table {{ width: 100%; border-collapse: collapse; }}
        th, td {{ padding: 8px; text-align: left; border-bottom: 1px solid #ddd; }}
        th {{ background-color: #f8f9fa; }}
        tr:nth-child(even) {{ background-color: #f8f9fa; }}
        .width-estimate {{ color: #7f8c8d; font-size: 0.9em; }}
        .space-info {{ background: #d5f4e6; padding: 5px; border-radius: 3px; margin-left: 10px; }}
        .over-capacity {{ background: #f8d7da; }}
        .near-capacity {{ background: #fff3cd; }}
        .good-capacity {{ background: #d1ecf1; }}
    </style>
</head>
<body>
    <div class="container">
        <h1>📚 Library Organization Guide</h1>
        
        <div class="summary">
            <h2>Collection Summary</h2>
            <p><strong>Total Books:</strong> {len([item for item in all_items if item['item_type'] == 'book'])}</p>
            <p><strong>Total Video Games:</strong> {len([item for item in all_items if item['item_type'] == 'videogame'])}</p>
            <p><strong>Total Music CDs:</strong> {len([item for item in all_items if item['item_type'] == 'music'])}</p>
            <p><strong>Organization Date:</strong> {Path(__file__).stat().st_mtime if Path(__file__).exists() else 'Unknown'}</p>
        </div>

        <div class="summary">
            <h2>Space Utilization Overview</h2>
    """
    
    for shelf_id, shelf_info in ORGANIZATION.items():
        usage = space_usage.get(shelf_id, {})
        utilization = usage.get('utilization', 0)
        
        if utilization > 90:
            status_class = 'over-capacity'
            status = '⚠️ Over capacity'
        elif utilization > 75:
            status_class = 'near-capacity' 
            status = '⚡ Near capacity'
        else:
            status_class = 'good-capacity'
            status = '✅ Good capacity'
            
        dimensions = SHELF_DIMENSIONS[shelf_id]
        
        html_content += f"""
            <div class="space-info {status_class}">
                <strong>{shelf_id.replace('_', ' ').title()}:</strong> 
                {utilization:.1f}% utilization 
                ({usage.get('total_width', 0):.1f}" used of {dimensions['width'] * dimensions['rows']}" available)
                <span style="margin-left: 10px;">{status}</span>
            </div>
        """
    
    html_content += """
        </div>
    """
    
    # Generate shelf sections
    books = [item for item in all_items if item['item_type'] == 'book']
    
    for shelf_id, shelf_rows in ORGANIZATION.items():
        shelf_name = shelf_id.replace('_', ' ').title()
        dimensions = SHELF_DIMENSIONS[shelf_id]
        usage = space_usage.get(shelf_id, {})
        
        html_content += f"""
        <div class="shelf">
            <div class="shelf-header">
                {shelf_name} ({dimensions['width']}" × {dimensions['depth']}" × {dimensions['rows']} rows)
                - {usage.get('utilization', 0):.1f}% utilized
            </div>
        """
        
        for row_num, row_name in shelf_rows.items():
            row_usage = usage.get('rows', {}).get(row_num, {})
            row_books = row_usage.get('items', [])
            row_width = row_usage.get('width', 0)
            
            html_content += f"""
            <div class="row">
                <div class="row-header">
                    <span>Row {row_num}: {row_name}</span>
                    <span class="width-estimate">{row_width:.1f}" estimated width ({len(row_books)} books)</span>
                </div>
                <table>
                    <thead>
                        <tr>
                            <th>Title</th>
                            <th>Author</th>
                            <th>Pages</th>
                            <th>Est. Width</th>
                        </tr>
                    </thead>
                    <tbody>
            """
            
            for book in sorted(row_books, key=lambda x: x['title']):
                html_content += f"""
                        <tr>
                            <td>{html.escape(book['title'])}</td>
                            <td>{html.escape(book['author'])}</td>
                            <td>{book['pages'] if book['pages'] > 0 else 'Unknown'}</td>
                            <td class="width-estimate">{book['width']:.2f}"</td>
                        </tr>
                """
            
            html_content += """
                    </tbody>
                </table>
            </div>
            """
        
        html_content += "</div>"
    
    html_content += """
        <div style="margin-top: 30px; padding: 15px; background: #f8f9fa; border-radius: 5px;">
            <h3>Notes:</h3>
            <ul>
                <li>Width estimates are based on page count: ~0.5" for <100 pages, ~1.0" for 200-300 pages, ~1.5" for 500-750 pages</li>
                <li>Actual widths may vary based on paper type, binding, and publisher</li>
                <li>Consider leaving 10-15% space on each shelf for easy access and future additions</li>
                <li>Video games and music CDs are not included in this organization (separate storage recommended)</li>
            </ul>
        </div>
    </div>
</body>
</html>
    """
    
    return html_content

def main():
    """Main execution function"""
    print("🔍 Processing library CSV files...")
    
    # Find all CSV files
    csv_files = list(Path('.').glob('library_*.csv'))
    print(f"Found {len(csv_files)} CSV files")
    
    all_items = []
    
    # Process each CSV file
    for csv_file in csv_files:
        print(f"📖 Processing {csv_file}...")
        items = parse_csv_file(csv_file)
        all_items.extend(items)
        print(f"   Found {len(items)} items")
    
    print(f"📊 Total items: {len(all_items)}")
    
    # Get only books for organization
    books = [item for item in all_items if item['item_type'] == 'book']
    print(f"📚 Books to organize: {len(books)}")
    
    # Iteratively balance until all shelves are under capacity
    print("⚖️  Performing iterative capacity balancing...")
    space_usage, iterations = iterative_balance(books)
    
    print(f"\n🎯 Final utilization after {iterations} iterations:")
    for shelf_id, data in space_usage.items():
        util = data['utilization']
        status = "✅ Good" if util <= 100 else "⚠️ Over"
        print(f"   {shelf_id}: {util:.1f}% utilization {status}")
    
    # Generate HTML
    print("\n🌐 Generating HTML...")
    html_content = generate_html(all_items, space_usage)
    
    # Write HTML file
    output_file = Path('library_organization.html')
    with open(output_file, 'w', encoding='utf-8') as f:
        f.write(html_content)
    
    print(f"✅ Generated {output_file}")
    print(f"🚀 Open {output_file} in your browser to view the organization guide")

if __name__ == '__main__':
    main()